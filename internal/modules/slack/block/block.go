package block

import (
	"fmt"
	"net"
	"strings"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/modules"
	"github.com/nt0xa/sonar/internal/templates"
	"github.com/slack-go/slack"
)

// Build creates Slack blocks from a notification
func Build(n *modules.Notification, codeBlocks []string) ([]slack.Block, error) {
	host, _, err := net.SplitHostPort(n.Event.RemoteAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to split host port: %w", err)
	}

	blocks := make([]slack.Block, 0)

	// Header block
	blocks = append(blocks, slack.NewHeaderBlock(
		&slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: fmt.Sprintf("%s [%s] %s",
				getHeaderEmoji(n.Event.Protocol),
				n.Payload.Name,
				strings.ToUpper(n.Event.Protocol)),
		},
	))

	// IP and Time section
	ipField := &slack.TextBlockObject{
		Type: slack.MarkdownType,
		Text: fmt.Sprintf(":satellite_antenna: *IP*\n%s", host),
	}

	timeField := &slack.TextBlockObject{
		Type: slack.MarkdownType,
		Text: fmt.Sprintf(":calendar: *Time*\n%s", n.Event.ReceivedAt.Format("02 Jan 2006 15:04:05 MST")),
	}

	blocks = append(blocks, slack.NewSectionBlock(
		nil,
		[]*slack.TextBlockObject{ipField, timeField},
		nil,
	))

	// GeoIP information if available
	if geoip := n.Event.Meta.GeoIP; geoip != nil {
		location := fmt.Sprintf(":round_pushpin: *Location*\n%s %s",
			templates.FlagEmoji(geoip.Country.ISOCode),
			geoip.Country.Name,
		)

		if geoip.City != "" {
			location += ", " + geoip.City
		}

		locationField := &slack.TextBlockObject{
			Type: slack.MarkdownType,
			Text: location,
		}

		orgField := &slack.TextBlockObject{
			Type: slack.MarkdownType,
			Text: fmt.Sprintf(":office: *Org*\n%s (AS%d)",
				geoip.ASN.Org,
				geoip.ASN.Number,
			),
		}

		blocks = append(blocks, slack.NewSectionBlock(
			nil,
			[]*slack.TextBlockObject{locationField, orgField},
			nil,
		))
	}

	// Email metadata if available
	if n.Event.Meta.SMTP != nil {
		email := n.Event.Meta.SMTP.Email
		var fromField, subjectField *slack.TextBlockObject

		if len(email.From) > 0 {
			var emails []string
			for _, f := range email.From {
				emails = append(emails, f.Address)
			}

			fromField = &slack.TextBlockObject{
				Type: slack.MarkdownType,
				Text: fmt.Sprintf(":bust_in_silhouette: *From*\n%s", strings.Join(emails, "\n")),
			}
		}

		if email.Subject != "" {
			subjectField = &slack.TextBlockObject{
				Type: slack.MarkdownType,
				Text: fmt.Sprintf(":memo: *Subject*\n%s", email.Subject),
			}
		}

		if fromField != nil || subjectField != nil {
			fields := make([]*slack.TextBlockObject, 0)
			if fromField != nil {
				fields = append(fields, fromField)
			}
			if subjectField != nil {
				fields = append(fields, subjectField)
			}

			blocks = append(blocks, slack.NewSectionBlock(
				nil,
				fields,
				nil,
			))
		}
	}

	for _, blockContent := range codeBlocks {
		blocks = append(blocks, slack.NewSectionBlock(
			&slack.TextBlockObject{
				Type: slack.MarkdownType,
				Text: fmt.Sprintf("```\n%s```", blockContent),
			},
			nil,
			nil,
		))
	}

	return blocks, nil
}

// getHeaderEmoji returns the appropriate emoji and color for the protocol
func getHeaderEmoji(protocol string) string {
	switch database.ProtoToCategory(protocol) {
	case database.ProtoCategoryDNS:
		return ":mag:"
	case database.ProtoCategoryFTP:
		return ":file_folder:"
	case database.ProtoCategorySMTP:
		return ":e-mail:"
	case database.ProtoCategoryHTTP:
		return ":globe_with_meridians:"
	default:
		return ":bell:"
	}
}
