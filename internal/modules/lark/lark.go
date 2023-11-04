package lark

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/core/httpserverext"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"

	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"

	"github.com/russtone/sonar/internal/actions"
	"github.com/russtone/sonar/internal/actionsdb"
	"github.com/russtone/sonar/internal/cmd"
	"github.com/russtone/sonar/internal/database"
	"github.com/russtone/sonar/internal/database/models"
	"github.com/russtone/sonar/internal/templates"
	"github.com/russtone/sonar/internal/utils/errors"
)

type Lark struct {
	db      *database.DB
	cfg     *Config
	cmd     *cmd.Command
	actions actions.Actions
	client  *lark.Client
	tls     *tls.Config
	tmpl    *templates.Templates

	domain string
}

func New(cfg *Config, db *database.DB, tlsConfig *tls.Config, acts actions.Actions, domain string) (*Lark, error) {

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
		lark.WithLogLevel(larkcore.LogLevelDebug),
		lark.WithHttpClient(httpClient))

	// Check that AppID and AppSecret are valid
	if resp, err := client.GetTenantAccessTokenBySelfBuiltApp(
		context.Background(),
		&larkcore.SelfBuiltTenantAccessTokenReq{AppID: cfg.AppID, AppSecret: cfg.AppSecret}); err != nil {
		return nil, fmt.Errorf("lark: %w", err)
	} else if !resp.Success() {
		return nil, fmt.Errorf("lark: invalid app id or app secret")
	}

	tmpl := templates.New(domain,
		templates.Markup(
			templates.Bold("**", "**")),
	)

	lrk := &Lark{
		client:  client,
		db:      db,
		cfg:     cfg,
		domain:  domain,
		actions: acts,
		tls:     tlsConfig,
		tmpl:    tmpl,
	}

	lrk.cmd = cmd.New(
		acts,
		cmd.PreExec(cmd.DefaultMessengersPreExec),
	)

	return lrk, nil
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

				user, err := lrk.db.UsersGetByParam(models.UserLarkID, *userID)

				if user == nil {
					// Create user if not exists
					user = &models.User{
						Name: fmt.Sprintf("user-%s", *userID),
						Params: models.UserParams{
							LarkUserID: *userID,
						},
					}

					if err := lrk.db.UsersCreate(user); err != nil {
						// TODO: logging
						lrk.mdMessage(*userID, msgID, "internal error")
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

				out, err := lrk.cmd.ParseAndExec(ctx, msg.Text, func(res actions.Result) error {
					s, err := lrk.tmpl.Execute(res)
					if err != nil {
						return err
					}
					lrk.mdMessage("", msgID, s)
					return nil
				})

				if out != "" {
					lrk.txtMessage("", msgID, out)
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
	lrk.txtMessage(userID, msgID, err.Error())
}

func (lrk *Lark) txtMessage(userID string, msgID *string, txt string) {
	lrk.cardMessage(userID, msgID, []*larkcard.MessageCardField{larkcard.NewMessageCardField().
		Text(larkcard.NewMessageCardPlainText().
			Content(txt).
			Build()).
		Build()})
}

func (lrk *Lark) mdMessage(userID string, msgID *string, md string) {
	lrk.cardMessage(userID, msgID, []*larkcard.MessageCardField{larkcard.NewMessageCardField().
		Text(larkcard.NewMessageCardLarkMd().
			Content(md).
			Build()).
		Build()})
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

func (lrk *Lark) cardMessage(userID string, msgID *string, fields []*larkcard.MessageCardField) {
	config := larkcard.NewMessageCardConfig().
		WideScreenMode(true).
		EnableForward(true).
		UpdateMulti(false).
		Build()

	// Elements
	div := larkcard.NewMessageCardDiv().
		Fields(fields).
		Build()

	card := larkcard.NewMessageCard().
		Config(config).
		Elements([]larkcard.MessageCardElement{div}).
		Build()

	content, err := card.String()
	if err != nil {
		// TODO: logging
		log.Println(err)
		return
	}

	lrk.sendMessage(userID, msgID, content)
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
		Parse(s),
	)
}
