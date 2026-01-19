package lark

import (
	"context"
	"fmt"
	"regexp"
	"unicode/utf8"

	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/modules"
	cardv2 "github.com/nt0xa/sonar/internal/modules/lark/card/v2"
)

// https://open.larksuite.com/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/message/create#:~:text=The%20maximum%20size%20of%20the,request%20body%20is%20150%20KB.

var emailRegexp = regexp.MustCompile(`(?i)([A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,24})`)

func (lrk *Lark) Name() string {
	return "lark"
}

func (lrk *Lark) Notify(ctx context.Context, n *modules.Notification) error {
	body := string(n.Event.RW)

	if n.Event.Protocol.Category() == models.ProtoCategorySMTP && n.Event.Meta.SMTP != nil {
		if text := n.Event.Meta.SMTP.Email.Text; text != "" {
			body = text
		}
	}

	// TODO: size limit
	if utf8.ValidString(body) {
		card, err := cardv2.Build(n, []byte(body))
		if err != nil {
			return fmt.Errorf("failed to build card: %w", err)
		}

		lrk.sendMessage(ctx, n.User.Params.LarkUserID, nil, string(card))
	} else {
		lrk.docMessage(ctx, n.User.Params.LarkUserID,
			fmt.Sprintf("log-%s-%s.txt", n.Payload.Name, n.Event.ReceivedAt.Format("15-04-05_02-Jan-2006")),
			"", []byte(body))
	}

	// For SMTP send mail.eml for better preview.
	if n.Event.Protocol.Category() == models.ProtoCategorySMTP && n.Event.Meta.SMTP != nil {
		data := n.Event.Meta.SMTP.Session.Data
		if data != "" {
			lrk.docMessage(ctx, n.User.Params.LarkUserID,
				fmt.Sprintf("mail-%s-%s.eml", n.Payload.Name, n.Event.ReceivedAt.Format("15-04-05_02-Jan-2006")),
				"", []byte(data))

			lrk.docMessage(ctx, n.User.Params.LarkUserID,
				fmt.Sprintf("mail-%s-%s.txt", n.Payload.Name, n.Event.ReceivedAt.Format("15-04-05_02-Jan-2006")),
				"", n.Event.RW)
		}
	}

	return nil
}
