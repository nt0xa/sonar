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
	"strings"
	"sync"
	"time"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/core/httpserverext"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"

	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/actionsdb"
	"github.com/nt0xa/sonar/internal/cmd"
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/templates"
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

	tmpl := templates.New(domain,
		templates.Default(
			templates.HTMLEscape(false),
			templates.Markup(
				templates.Bold("**", "**"),
				templates.CodeBlock("```", "```"),
			),
		),
		// Disable markup for notification header.
		templates.PerTemplate(templates.NotificationHeaderID,
			templates.HTMLEscape(false),
			templates.Markup(),
		),
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
	var dispatcher *dispatcher.EventDispatcher

	// Webhooks by default
	if lrk.cfg.Mode == ModeWebhook || lrk.cfg.Mode == "" {
		dispatcher = lrk.makeDispatcher(lrk.cfg.VerificationToken, lrk.cfg.EncryptKey, true)
		return lrk.startWebhook(dispatcher)
	} else {
		dispatcher = lrk.makeDispatcher("", "", false)
		return lrk.startWebsocket(dispatcher)
	}
}

func (lrk *Lark) makeDispatcher(verificationToken, eventEncryptKey string, dedupEvents bool) *dispatcher.EventDispatcher {
	// Sometimes the same event is sent several times, so keep recent event ids
	// to prevent handling the same event more than once.
	recentEvents := map[string]time.Time{}
	recentEventsMutex := sync.Mutex{}

	if dedupEvents {
		go func() {
			ticker := time.NewTicker(10 * time.Second)
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
	}

	return dispatcher.NewEventDispatcher(verificationToken, eventEncryptKey).
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

				if dedupEvents {
					eventID := event.EventV2Base.Header.EventID

					recentEventsMutex.Lock()
					if _, ok := recentEvents[eventID]; ok {
						recentEventsMutex.Unlock()

						// Event was already handled
						return nil
					}
					recentEvents[eventID] = time.Now()
					recentEventsMutex.Unlock()
				}

				msgID := event.Event.Message.MessageId

				if event.Event.Message.ChatType == nil {
					fmt.Println("chat type is nil")
					return nil
				}

				var userID *string
				switch *event.Event.Message.ChatType {
				case "p2p":
					userID = event.Event.Sender.SenderId.UserId
				case "group":
					userID = event.Event.Message.ChatId
				default:
					fmt.Println("unknown chat type")
					return nil
				}

				if userID == nil || msgID == nil {
					// TODO: better error
					return nil
				}

				user, err := lrk.db.UsersGetByParam(ctx, models.UserLarkID, *userID)

				if user == nil {
					// Create user if not exists
					user = &models.User{
						Name: fmt.Sprintf("user-%s", *userID),
						Params: models.UserParams{
							LarkUserID: *userID,
						},
					}

					if err := lrk.db.UsersCreate(ctx, user); err != nil {
						// TODO: logging
						lrk.message(*userID, msgID, "internal error")
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

				text := msg.Text

				// remove @mention from the text
				if event.Event.Message.Mentions != nil {
					for _, mention := range event.Event.Message.Mentions {
						text = strings.ReplaceAll(text, *mention.Key, "")
					}
				}

				text = strings.TrimSpace(text)

				stdout, stderr, err := lrk.cmd.ParseAndExec(ctx, text, func(res actions.Result) error {
					s, err := lrk.tmpl.RenderResult(res)
					if err != nil {
						return err
					}

					lrk.message("", msgID, s)

					return nil
				})
				if err != nil {
					lrk.message("", msgID, err.Error())
					return nil
				}

				if stdout != "" {
					lrk.message("", msgID, stdout)
				}

				if stderr != "" {
					lrk.message("", msgID, stderr)
				}

				return nil
			},
		)
}

func (lrk *Lark) startWebsocket(eventHandler *dispatcher.EventDispatcher) error {
	cli := larkws.NewClient(lrk.cfg.AppID, lrk.cfg.AppSecret,
		larkws.WithEventHandler(eventHandler),
		larkws.WithLogLevel(larkcore.LogLevelInfo),
	)

	return cli.Start(context.TODO())
}

func (lrk *Lark) startWebhook(eventHandler *dispatcher.EventDispatcher) error {
	mux := http.NewServeMux()

	// TODO: take path from config
	mux.HandleFunc("/webhook/event", httpserverext.NewEventHandlerFunc(eventHandler,
		larkevent.WithLogLevel(larkcore.LogLevelInfo)))

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

func (lrk *Lark) message(userID string, msgID *string, content string) {
	config := larkcard.NewMessageCardConfig().
		WideScreenMode(true).
		EnableForward(true).
		UpdateMulti(false).
		Build()

	md := larkcard.NewMessageCardMarkdown().
		Content(content).
		Build()

	card := larkcard.NewMessageCard().
		Config(config).
		Elements([]larkcard.MessageCardElement{md}).
		Build()

	content, err := card.String()
	if err != nil {
		// TODO: logging
		log.Println(err)
		return
	}

	lrk.sendMessage(userID, msgID, content)
}

func (lrk *Lark) sendMessage(userID string, msgID *string, content string) {
	var err error

	if msgID != nil {
		_, err = lrk.client.Im.Message.Reply(context.Background(), larkim.NewReplyMessageReqBuilder().
			MessageId(*msgID).
			Body(larkim.NewReplyMessageReqBodyBuilder().
				MsgType(larkim.MsgTypeInteractive).
				Content(content).
				Build()).
			Build())
	} else {
		idType := larkim.ReceiveIdTypeUserId

		if strings.HasPrefix(userID, "oc_") {
			idType = larkim.ReceiveIdTypeChatId
		}

		_, err = lrk.client.Im.Message.Create(context.Background(), larkim.NewCreateMessageReqBuilder().
			ReceiveIdType(idType).
			Body(larkim.NewCreateMessageReqBodyBuilder().
				MsgType(larkim.MsgTypeInteractive).
				ReceiveId(userID).
				Content(content).
				Build()).
			Build())
	}

	if err != nil {
		fmt.Println(err)
	}
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

	idType := larkim.ReceiveIdTypeUserId

	if strings.HasPrefix(chatID, "oc_") {
		idType = larkim.ReceiveIdTypeChatId
	}

	resp2, err := lrk.client.Im.Message.Create(context.Background(), larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(idType).
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
