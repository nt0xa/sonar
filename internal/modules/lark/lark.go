package lark

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"text/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/google/shlex"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/core/httpserverext"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	larkcontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"

	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	"github.com/spf13/cobra"

	"github.com/russtone/sonar/internal/actions"
	"github.com/russtone/sonar/internal/actionsdb"
	"github.com/russtone/sonar/internal/cmd"
	"github.com/russtone/sonar/internal/database"
	"github.com/russtone/sonar/internal/database/models"
	"github.com/russtone/sonar/internal/utils/errors"
)

type Lark struct {
	db      *database.DB
	cfg     *Config
	cmd     cmd.Command
	actions actions.Actions
	client  *lark.Client
	tls     *tls.Config

	domain string
}

func New(cfg *Config, db *database.DB, tlsConfig *tls.Config, actions actions.Actions, domain string) (*Lark, error) {

	httpClient := http.DefaultClient

	// Proxy
	if cfg.ProxyURL != "" {
		proxyURL, err := url.Parse(cfg.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("lark: invalid proxy url: %w", err)
		}

		httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.ProxyInsecure,
			},
		}

	}

	// TODO: better logging
	var client = lark.NewClient(
		cfg.AppID,
		cfg.AppSecret,
		lark.WithLogReqAtDebug(true),
		lark.WithLogLevel(larkcore.LogLevelInfo),
		lark.WithHttpClient(httpClient))

	// Check that AppID and AppSecret are valid
	if resp, err := client.GetTenantAccessTokenBySelfBuiltApp(
		context.Background(),
		&larkcore.SelfBuiltTenantAccessTokenReq{AppID: cfg.AppID, AppSecret: cfg.AppSecret}); err != nil {
		return nil, fmt.Errorf("lark: %w", err)
	} else if !resp.Success() {
		return nil, fmt.Errorf("lark: invalid app id or app secret")
	}

	lrk := &Lark{
		client:  client,
		db:      db,
		cfg:     cfg,
		domain:  domain,
		actions: actions,
		tls:     tlsConfig,
	}

	lrk.cmd = cmd.New(actions, lrk, lrk.preExec)

	return lrk, nil
}

func (lrk *Lark) preExec(root *cobra.Command, u *actions.User) {
}

func (lrk *Lark) Start() error {

	// Sometimes the same event is sent several times, so keep recent event ids
	// to prevent handling the same event more than once.
	recentEvents := map[string]time.Time{}
	recentEventsMutex := sync.Mutex{}

	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				toRemove := make([]string, 0)

				for eventID, handledAt := range recentEvents {

					// Cleanup events after 10m
					// TODO: config
					if time.Since(handledAt) > time.Minute*5 {
						toRemove = append(toRemove, eventID)
					}
				}

				recentEventsMutex.Lock()
				for _, eventID := range toRemove {
					delete(recentEvents, eventID)
				}
				recentEventsMutex.Unlock()
			}
		}
	}()

	handler := dispatcher.NewEventDispatcher(lrk.cfg.VerificationToken, lrk.cfg.EncryptKey).
		OnP2MessageReceiveV1(
			func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {

				ts, err := strconv.ParseInt(*event.Event.Message.CreateTime, 10, 64)
				if err != nil {
					return nil
				}

				// Do not handle for old events
				// TODO: config
				if time.Since(time.UnixMilli(ts)) > time.Minute*5 {
					return nil
				}

				eventID := event.EventV2Base.Header.EventID

				recentEventsMutex.Lock()
				if _, ok := recentEvents[eventID]; ok {
					recentEventsMutex.Unlock()

					// Event was already handled
					return nil
				}
				recentEvents[eventID] = time.Now()
				recentEventsMutex.Unlock()

				userID := event.Event.Sender.SenderId.UserId
				msgID := event.Event.Message.MessageId

				if userID == nil || msgID == nil {
					// TODO: better error
					return nil
				}

				var user *models.User

				if lrk.cfg.DepartmentID != "" {
					// TODO: save last check date in user and don't request departement every time.
					resp, err := lrk.client.Contact.User.Get(ctx,
						larkcontact.NewGetUserReqBuilder().
							UserId(*userID).
							UserIdType(larkim.UserIdTypeUserId).
							Build())

					if err != nil {
						// TODO: logging
						fmt.Println(err)
						return nil
					}

					found := false
					for _, departmentID := range resp.Data.User.DepartmentIds {
						if departmentID == lrk.cfg.DepartmentID {
							found = true
							break
						}
					}

					if !found {
						// TODO: logging & better message
						lrk.mdMessage(*userID, msgID, "access denied")
						return nil
					}

					user, err = lrk.db.UsersGetByParam(models.UserLarkID, *userID)
					if user == nil && lrk.cfg.DepartmentID != "" {
						// Create user
						user = &models.User{
							Name: *resp.Data.User.Name,
							Params: models.UserParams{
								LarkUserID: *resp.Data.User.UserId,
							},
						}
						if err := lrk.db.UsersCreate(user); err != nil {
							// TODO: logging
							fmt.Println(err)
							return nil
						}
					}

				} else {
					user, err = lrk.db.UsersGetByParam(models.UserLarkID, *userID)
					if err != nil {
						// TODO: logging
						fmt.Println(err)
						return nil
					}
				}

				type TextMessage struct {
					Text string `json:"text"`
				}

				var msg TextMessage

				if err := json.Unmarshal([]byte(*event.Event.Message.Content), &msg); err != nil {
					// TODO: logging
					fmt.Println(err)
					return nil
				}

				ctx = SetMessageID(ctx, *msgID)
				ctx = actionsdb.SetUser(ctx, user)
				args, _ := shlex.Split(strings.TrimLeft(msg.Text, "/"))

				// It is important to pass false as "local" here to disable
				// dangerous commands.
				if out, err := lrk.cmd.Exec(ctx, actionsdb.User(user), false, args); err != nil {
					lrk.handleError(*userID, msgID, err)
				} else if out != "" {
					lrk.mdMessage(*userID, msgID, out)
				}

				return nil
			},
		)

	mux := http.NewServeMux()

	// TODO: take path from config
	mux.HandleFunc("/webhook/event", httpserverext.NewEventHandlerFunc(handler,
		larkevent.WithLogLevel(larkcore.LogLevelDebug)))

	// TODO: take port from config
	srv := http.Server{
		Addr:      ":31338",
		Handler:   mux,
		TLSConfig: lrk.tls,
	}

	if lrk.cfg.TLSEnabled {
		return srv.ListenAndServeTLS("", "")
	} else {
		return srv.ListenAndServe()
	}
}

// TODO: add common notifier module
// use something like capabilities to determine if current notifier supports specific features.

func (lrk *Lark) handleError(userID string, msgID *string, err errors.Error) {
	lrk.mdMessage(userID, msgID, err.Error())
}

func (lrk *Lark) txtMessage(userID string, msgID *string, txt string) {
	content := larkim.NewTextMsgBuilder().
		Text(txt).
		Build()

	resp, err := lrk.client.Im.Message.Create(context.Background(), larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(larkim.ReceiveIdTypeUserId).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			MsgType(larkim.MsgTypeText).
			ReceiveId(userID).
			Content(content).
			Build()).
		Build())

	fmt.Println(resp, err)
}

func (lrk *Lark) sendMessage(userID string, msgID *string, content string) {
	if msgID != nil {
		resp, err := lrk.client.Im.Message.Reply(context.Background(), larkim.NewReplyMessageReqBuilder().
			MessageId(*msgID).
			Body(larkim.NewReplyMessageReqBodyBuilder().
				MsgType(larkim.MsgTypeInteractive).
				Content(content).
				Build()).
			Build())
		fmt.Println(resp, err)
	} else {
		resp, err := lrk.client.Im.Message.Create(context.Background(), larkim.NewCreateMessageReqBuilder().
			ReceiveIdType(larkim.ReceiveIdTypeUserId).
			Body(larkim.NewCreateMessageReqBodyBuilder().
				MsgType(larkim.MsgTypeInteractive).
				ReceiveId(userID).
				Content(content).
				Build()).
			Build())
		fmt.Println(resp, err)
	}
}

func (lrk *Lark) cardMessage(userID string, msgID *string, card *larkcard.MessageCard) {
	content, err := card.String()
	if err != nil {
		log.Println(err)
	}

	lrk.sendMessage(userID, msgID, content)
}

func (lrk *Lark) mdMessage(userID string, msgID *string, md string) {
	config := larkcard.NewMessageCardConfig().
		WideScreenMode(true).
		EnableForward(true).
		UpdateMulti(false).
		Build()

	// Elements
	div := larkcard.NewMessageCardDiv().
		Fields([]*larkcard.MessageCardField{larkcard.NewMessageCardField().
			Text(larkcard.NewMessageCardLarkMd().
				Content(md).
				Build()).
			IsShort(true).
			Build()}).
		Build()

	card := larkcard.NewMessageCard().
		Config(config).
		Elements([]larkcard.MessageCardElement{div}).
		Build()

	lrk.cardMessage(userID, msgID, card)
}

func (lrk *Lark) docMessage(chatID string, name string, caption string, data []byte) {
	file := bytes.NewReader(data)

	resp, err := lrk.client.Im.File.Create(context.Background(),
		larkim.NewCreateFileReqBuilder().
			Body(larkim.NewCreateFileReqBodyBuilder().
				FileType(larkim.FileTypePdf).
				FileName(name).
				File(file).
				Build()).
			Build())

	if err != nil {
		log.Println(err)
		return
	}

	if !resp.Success() {
		log.Println(resp.Code, resp.Msg, resp.RequestId())
		return
	}

	msg := larkim.MessageFile{FileKey: *resp.Data.FileKey}
	content, err := msg.String()
	if err != nil {
		log.Println(err)
		return
	}

	resp2, err := lrk.client.Im.Message.Create(context.Background(), larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(larkim.ReceiveIdTypeUserId).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			MsgType(larkim.MsgTypeFile).
			ReceiveId(chatID).
			Content(content).
			Build()).
		Build())

	if err != nil {
		log.Println(err)
		return
	}

	if !resp2.Success() {
		log.Println(resp.Code, resp.Msg, resp.RequestId())
		return
	}

}

func tpl(s string) *template.Template {
	return template.Must(template.
		New("").
		Funcs(sprig.FuncMap()).
		Funcs(template.FuncMap{
			// This is nesessary for templates to compile.
			// It will be replaced later with correct function.
			"domain": func() string { return "" },
		}).
		Parse(s),
	)
}

func (lrk *Lark) getDomain() string {
	return lrk.domain
}

func (lrk *Lark) txtResult(ctx context.Context, txt string) {
	u, err := actionsdb.GetUser(ctx)
	if err != nil {
		return
	}

	if msgID, err := GetMessageID(ctx); err == nil && msgID != nil {
		lrk.mdMessage(u.Params.LarkUserID, msgID, txt)
	} else {
		lrk.mdMessage(u.Params.LarkUserID, nil, txt)
	}
}

func (lrk *Lark) tplResult(ctx context.Context, tpl *template.Template, data interface{}) {
	tpl.Funcs(template.FuncMap{
		"domain": lrk.getDomain,
	})

	buf := &bytes.Buffer{}

	if err := tpl.Execute(buf, data); err != nil {
		// lrk.handleError(u.Params.LarkUserID, nil, errors.Internal(err))
		log.Println(err)
	}

	lrk.txtResult(ctx, buf.String())
}
