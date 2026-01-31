package block_test

import (
	"testing"
	"time"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/modules"
	"github.com/nt0xa/sonar/internal/modules/slack/block"
	"github.com/nt0xa/sonar/pkg/geoipx"
	"github.com/nt0xa/sonar/pkg/smtpx"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/require"
)

func TestBuild(t *testing.T) {
	receivedAt, _ := time.Parse("2006-01-02T15:04:05Z", "2023-01-01T00:00:00Z")

	blocks, err := block.Build(&modules.Notification{
		User:    &database.User{},
		Payload: &database.Payload{Name: "test"},
		Event: &database.Event{
			Protocol: "http",
			RW:       []byte("test"),
			Meta: database.EventsMeta{
				GeoIP: &geoipx.Meta{
					City: "London",
					Country: geoipx.Country{
						Name:    "United Kingdom",
						ISOCode: "GB",
					},
					ASN: geoipx.ASN{
						Org:    "Google Inc.",
						Number: 1234,
					},
				},
			},
			RemoteAddr: "10.13.37.1:1337",
			ReceivedAt: receivedAt,
		},
	}, []string{"test"})

	require.NoError(t, err)
	require.NotNil(t, blocks)

	// Verify we have the expected number of blocks
	// 1. Header
	// 2. IP/Time section
	// 3. Location/Org section
	// 4. Divider
	// 5. Request/Response section
	require.Len(t, blocks, 4)

	// Verify header block
	headerBlock, ok := blocks[0].(*slack.HeaderBlock)
	require.True(t, ok)
	require.Contains(t, headerBlock.Text.Text, "[test]")
	require.Contains(t, headerBlock.Text.Text, "HTTP")

	// Verify IP/Time section
	ipTimeBlock, ok := blocks[1].(*slack.SectionBlock)
	require.True(t, ok)
	require.Len(t, ipTimeBlock.Fields, 2)
	require.Contains(t, ipTimeBlock.Fields[0].Text, "10.13.37.1")
	require.Contains(t, ipTimeBlock.Fields[1].Text, "01 Jan 2023")

	// Verify Location/Org section
	geoBlock, ok := blocks[2].(*slack.SectionBlock)
	require.True(t, ok)
	require.Len(t, geoBlock.Fields, 2)
	require.Contains(t, geoBlock.Fields[0].Text, "United Kingdom")
	require.Contains(t, geoBlock.Fields[0].Text, "London")
	require.Contains(t, geoBlock.Fields[1].Text, "Google Inc.")
	require.Contains(t, geoBlock.Fields[1].Text, "AS1234")

	// Verify request/response section
	rwBlock, ok := blocks[3].(*slack.SectionBlock)
	require.True(t, ok)
	require.Contains(t, rwBlock.Text.Text, "test")
}

func TestBuildMinimal(t *testing.T) {
	receivedAt, _ := time.Parse("2006-01-02T15:04:05Z", "2023-01-01T00:00:00Z")

	blocks, err := block.Build(&modules.Notification{
		User:    &database.User{},
		Payload: &database.Payload{Name: "test"},
		Event: &database.Event{
			Protocol:   "dns",
			RW:         []byte("test"),
			Meta:       database.EventsMeta{},
			RemoteAddr: "10.13.37.1:1337",
			ReceivedAt: receivedAt,
		},
	}, []string{"test"})

	require.NoError(t, err)
	require.NotNil(t, blocks)

	// Verify we have the expected number of blocks without geoip/email
	// 1. Header
	// 2. IP/Time section
	// 3. Divider
	// 4. Request/Response section
	require.Len(t, blocks, 3)

	// Verify header block uses DNS emoji
	headerBlock, ok := blocks[0].(*slack.HeaderBlock)
	require.True(t, ok)
	require.Contains(t, headerBlock.Text.Text, "DNS")
}

func TestBuildWithEmail(t *testing.T) {
	receivedAt, _ := time.Parse("2006-01-02T15:04:05Z", "2023-01-01T00:00:00Z")

	blocks, err := block.Build(&modules.Notification{
		User:    &database.User{},
		Payload: &database.Payload{Name: "test"},
		Event: &database.Event{
			Protocol: "smtp",
			RW:       []byte("test"),
			Meta: database.EventsMeta{
				SMTP: &smtpx.Meta{
					Email: smtpx.Email{
						From: []smtpx.Address{
							{Address: "sender@example.com"},
						},
						Subject: "Test Subject",
					},
				},
			},
			RemoteAddr: "10.13.37.1:1337",
			ReceivedAt: receivedAt,
		},
	}, []string{"test"})

	require.NoError(t, err)
	require.NotNil(t, blocks)

	// Should have email section
	require.Len(t, blocks, 4)

	// Verify email section
	emailBlock, ok := blocks[2].(*slack.SectionBlock)
	require.True(t, ok)
	require.Len(t, emailBlock.Fields, 2)
	require.Contains(t, emailBlock.Fields[0].Text, "sender@example.com")
	require.Contains(t, emailBlock.Fields[1].Text, "Test Subject")
}
