package slack

import (
	"context"
	"fmt"

	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/modules"
	"github.com/nt0xa/sonar/internal/modules/slack/block"
	"github.com/slack-go/slack"
)

func (s *Slack) Name() string {
	return "slack"
}

func (s *Slack) Notify(ctx context.Context, n *modules.Notification) error {
	codeBlocks := make([]string, 0)

	if n.Event.Protocol.Category() == models.ProtoCategorySMTP && n.Event.Meta.SMTP != nil {
		if text := n.Event.Meta.SMTP.Email.Text; len(text) > 0 {
			codeBlocks = append(codeBlocks, text)
		}
	} else {
		codeBlocks = append(codeBlocks, string(n.Event.RW))
	}

	blocks, err := block.Build(n, codeBlocks)
	if err != nil {
		return fmt.Errorf("failed to build blocks: %w", err)
	}

	channelID, timestamp, err := s.client.PostMessageContext(
		ctx,
		n.User.Params.SlackID,
		slack.MsgOptionBlocks(blocks...),
	)
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}

	// For SMTP send mail.eml for better preview
	if n.Event.Protocol.Category() == models.ProtoCategorySMTP && n.Event.Meta.SMTP != nil {
		data := n.Event.Meta.SMTP.Session.Data

		// Upload .eml file
		if len(data) > 0 {
			_, err := s.client.UploadFileV2Context(ctx, slack.UploadFileV2Parameters{
				Channel:         channelID,
				ThreadTimestamp: timestamp,
				Filename:        fmt.Sprintf("email-%s-%s.eml", n.Payload.Name, n.Event.ReceivedAt.Format("15-04-05_02-Jan-2006")),
				FileSize:        len(data),
				Content:         data,
			})
			if err != nil {
				return fmt.Errorf("failed to upload eml file: %w", err)
			}
		}

		// Upload .txt file
		if len(n.Event.RW) >= 0 {
			_, err := s.client.UploadFileV2Context(ctx, slack.UploadFileV2Parameters{
				Channel:         channelID,
				ThreadTimestamp: timestamp,
				Filename:        fmt.Sprintf("smtp-%s-%s.txt", n.Payload.Name, n.Event.ReceivedAt.Format("15-04-05_02-Jan-2006")),
				FileSize:        len(n.Event.RW),
				Content:         string(n.Event.RW),
			})
			if err != nil {
				return fmt.Errorf("failed to upload txt file: %w", err)
			}
		}
	}

	return nil
}
