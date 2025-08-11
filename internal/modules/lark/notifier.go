package lark

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/modules"
	cardv2 "github.com/nt0xa/sonar/internal/modules/lark/card/v2"
)

// https://open.larksuite.com/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/message/create#:~:text=The%20maximum%20size%20of%20the,request%20body%20is%20150%20KB.

var emailRegexp = regexp.MustCompile("(?i)([A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,24})")

func (lrk *Lark) Name() string {
	return "lark"
}

func (lrk *Lark) Notify(ctx context.Context, n *modules.Notification) error {
	body := string(n.Event.RW)

	// TODO: Change when .Meta is struct
	if n.Event.Protocol.Category() == models.ProtoCategorySMTP {
		if email, ok := n.Event.Meta["email"].(map[string]any); ok {
			if text := email["text"]; text != nil {
				body = text.(string)
			}
		}
	}

	// Bypass: msg:The messages do NOT pass the audit, ext=contain sensitive data: EMAIL_ADDRESS,code:230028
	for _, m := range emailRegexp.FindAllString(body, -1) {
		body = strings.ReplaceAll(body, m, strings.Replace(m, "@", "＠", 1))
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
	if n.Event.Protocol.Category() == models.ProtoCategorySMTP {
		sess, ok := n.Event.Meta["session"].(map[string]interface{})
		if !ok {
			return nil
		}
		data, ok := sess["data"].(string)
		if !ok {
			return nil
		}

		lrk.docMessage(ctx, n.User.Params.LarkUserID,
			fmt.Sprintf("mail-%s-%s.eml", n.Payload.Name, n.Event.ReceivedAt.Format("15-04-05_02-Jan-2006")),
			"", []byte(data))

		lrk.docMessage(ctx, n.User.Params.LarkUserID,
			fmt.Sprintf("mail-%s-%s.txt", n.Payload.Name, n.Event.ReceivedAt.Format("15-04-05_02-Jan-2006")),
			"", n.Event.RW)
	}

	return nil
}
