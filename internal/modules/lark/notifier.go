package lark

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/modules"
)

// https://open.larksuite.com/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/message/create#:~:text=The%20maximum%20size%20of%20the,request%20body%20is%20150%20KB.

var emailRegexp = regexp.MustCompile("(?i)([A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,24})")

func (lrk *Lark) Notify(n *modules.Notification) error {
	header, body, err := lrk.tmpl.RenderNotification(n)
	if err != nil {
		return err
	}

	for _, m := range emailRegexp.FindAllString(body, -1) {
		body = strings.ReplaceAll(body, m, strings.Replace(m, "@", "＠", 1))
	}

	body = strings.ReplaceAll(body, "<img", "＜img")

	config := larkcard.NewMessageCardConfig().
		WideScreenMode(true).
		EnableForward(true).
		UpdateMulti(false).
		Build()

		// header
	var template string

	switch n.Event.Protocol.Category() {
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
			Content(header).
			Build()).
		Build()

	if utf8.ValidString(body) {

		// Elements
		md := larkcard.NewMessageCardMarkdown().
			Content(body).
			Build()

		card := larkcard.NewMessageCard().
			Config(config).
			Header(cardHeader).
			Elements([]larkcard.MessageCardElement{md}).
			Build()

		content, err := card.String()
		if err != nil {
			// TODO: logging
			return fmt.Errorf("lark: %w", err)
		}

		lrk.sendMessage(n.User.Params.LarkUserID, nil, content)
	} else {
		lrk.docMessage(n.User.Params.LarkUserID,
			fmt.Sprintf("log-%s-%s.txt", n.Payload.Name, n.Event.ReceivedAt.Format("15-04-05_02-Jan-2006")),
			header, n.Event.RW)
	}

	// For SMTP send mail.eml for better preview.
	if n.Event.Protocol.Category() == models.ProtoCategorySMTP {
		sess, ok := n.Event.Meta["session"].(map[string]interface{})
		if !ok {
			return nil
		}
		data, ok := sess["data"].(string)
		if !ok {
			return nil
		}
		lrk.docMessage(n.User.Params.LarkUserID,
			fmt.Sprintf("mail-%s-%s.eml", n.Payload.Name, n.Event.ReceivedAt.Format("15-04-05_02-Jan-2006")),
			header, []byte(data))
	}

	return nil
}
