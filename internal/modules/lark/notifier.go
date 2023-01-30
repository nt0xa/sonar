package lark

import (
	"bytes"
	"fmt"
	"unicode/utf8"

	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	"github.com/russtone/sonar/internal/database/models"
)

// https://open.larksuite.com/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/message/create#:~:text=The%20maximum%20size%20of%20the,request%20body%20is%20150%20KB.

var (
	messageHeaderTemplate = tpl(`[{{ .Name }}] {{ .Protocol | upper }} from {{ .RemoteAddr }} at {{ .ReceivedAt }}`)
	messageBodyTemplate   = tpl(`
{{- if eq .Protocol "smtp" }}
**Rcpt To:** {{ index (index .Meta "session") "rcptTo" | join ", " }}
**Mail From:** {{ index (index .Meta "session") "mailFrom" | join ", " }}
{{ end }}
{{ .Data }}
`)
)

func (lrk *Lark) Notify(u *models.User, p *models.Payload, e *models.Event) error {

	var header, body bytes.Buffer

	headerData := struct {
		Name       string
		Protocol   string
		RemoteAddr string
		ReceivedAt string
		Meta       map[string]interface{}
	}{
		p.Name,
		e.Protocol.String(),
		e.RemoteAddr,
		e.ReceivedAt.Format("15:04:05 on 02 Jan 2006"),
		e.Meta,
	}

	if err := messageHeaderTemplate.Execute(&header, headerData); err != nil {
		fmt.Println(err)
		return fmt.Errorf("message header render error: %w", err)
	}

	bodyData := struct {
		Data     string
		Protocol string
		Meta     map[string]interface{}
	}{
		string(e.RW),
		e.Protocol.String(),
		e.Meta,
	}

	if err := messageBodyTemplate.Execute(&body, bodyData); err != nil {
		return fmt.Errorf("message body render error: %w", err)
	}

	config := larkcard.NewMessageCardConfig().
		WideScreenMode(true).
		EnableForward(true).
		UpdateMulti(false).
		Build()

		// header
	var template string

	switch e.Protocol.Category() {
	case models.ProtoCategoryDNS:
		template = larkcard.TemplateCarmine
	case models.ProtoCategoryFTP:
		template = larkcard.TemplateTurquoise
	case models.ProtoCategorySMTP:
		template = larkcard.TemplateIndigo
	case models.ProtoCategoryHTTP:
		template = larkcard.TemplateWathet
	}

	cardHeader := larkcard.NewMessageCardHeader().
		Template(template).
		Title(larkcard.NewMessageCardPlainText().
			Content(header.String()).
			Build()).
		Build()

	if utf8.ValidString(body.String()) {

		// Elements
		div := larkcard.NewMessageCardDiv().
			Fields([]*larkcard.MessageCardField{larkcard.NewMessageCardField().
				Text(larkcard.NewMessageCardLarkMd().
					Content(body.String()).
					Build()).
				IsShort(true).
				Build()}).
			Build()

		card := larkcard.NewMessageCard().
			Config(config).
			Header(cardHeader).
			Elements([]larkcard.MessageCardElement{div}).
			Build()
		lrk.cardMessage(u.Params.LarkUserID, nil, card)
	} else {
		lrk.docMessage(u.Params.LarkUserID,
			fmt.Sprintf("log-%s-%s.txt", p.Name, e.ReceivedAt.Format("15-04-05_02-Jan-2006")),
			header.String(), e.RW)
	}

	// For SMTP send mail.eml for better preview.
	if e.Protocol.Category() == models.ProtoCategorySMTP {
		sess, ok := e.Meta["session"].(map[string]interface{})
		if !ok {
			return nil
		}
		data, ok := sess["data"].(string)
		if !ok {
			return nil
		}
		lrk.docMessage(u.Params.LarkUserID,
			fmt.Sprintf("mail-%s-%s.eml", p.Name, e.ReceivedAt.Format("15-04-05_02-Jan-2006")),
			header.String(), []byte(data))
	}

	return nil
}
