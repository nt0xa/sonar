package block

import (
	"fmt"
	"net"
	"strings"

	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/modules"
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
				strings.ToUpper(n.Event.Protocol.String())),
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
	if geo, ok := n.Event.Meta["geoip"]; ok {
		g, _ := geo.(map[string]any)
		var locationField, orgField *slack.TextBlockObject

		if country, ok := g["country"].(map[string]any); ok {
			location := fmt.Sprintf(":round_pushpin: *Location*\n%s %s",
				country["flagEmoji"],
				country["name"],
			)

			if city, ok := g["city"]; ok {
				location += ", " + city.(string)
			}

			locationField = &slack.TextBlockObject{
				Type: slack.MarkdownType,
				Text: location,
			}
		}

		if asn, ok := g["asn"].(map[string]any); ok {
			orgField = &slack.TextBlockObject{
				Type: slack.MarkdownType,
				Text: fmt.Sprintf(":office: *Org*\n%s (AS%d)",
					asn["org"],
					asn["number"],
				),
			}
		}

		if locationField != nil || orgField != nil {
			fields := make([]*slack.TextBlockObject, 0)
			if locationField != nil {
				fields = append(fields, locationField)
			}
			if orgField != nil {
				fields = append(fields, orgField)
			}

			blocks = append(blocks, slack.NewSectionBlock(
				nil,
				fields,
				nil,
			))
		}
	}

	// Email metadata if available
	if email, ok := n.Event.Meta["email"].(map[string]any); ok {
		var fromField, subjectField *slack.TextBlockObject

		if from, ok := email["from"].([]any); ok {
			var emails []string
			for _, f := range from {
				emails = append(emails, f.(map[string]any)["email"].(string))
			}

			fromField = &slack.TextBlockObject{
				Type: slack.MarkdownType,
				Text: fmt.Sprintf(":bust_in_silhouette: *From*\n%s", strings.Join(emails, "\n")),
			}
		}

		if subject, ok := email["subject"].(string); ok {
			subjectField = &slack.TextBlockObject{
				Type: slack.MarkdownType,
				Text: fmt.Sprintf(":memo: *Subject*\n%s", subject),
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
func getHeaderEmoji(protocol models.Proto) string {
	switch protocol.Category() {
	case models.ProtoCategoryDNS:
		return ":mag:"
	case models.ProtoCategoryFTP:
		return ":file_folder:"
	case models.ProtoCategorySMTP:
		return ":e-mail:"
	case models.ProtoCategoryHTTP:
		return ":globe_with_meridians:"
	default:
		return ":bell:"
	}
}
