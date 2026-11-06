package block_test

import (
	"strings"
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

	blocks, text, err := block.Build(&modules.Notification{
		User:    &database.User{},
		Payload: &database.Payload{Name: "test"},
		Event: &database.Event{
			Protocol: "http",
			Data:     [][]byte{[]byte("test")},
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
	require.Equal(t, "test", codeBlockText(t, blocks[3]))

	require.Equal(t, "[test] HTTP from 10.13.37.1", text)
}

func TestBuildCodeBlock(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"short", "abc", "abc"},
		{"exact", strings.Repeat("a", 3000), strings.Repeat("a", 3000)},
		{"long", strings.Repeat("a", 3001), strings.Repeat("a", 2997) + "..."},
		{"multibyte", strings.Repeat("я", 3001), strings.Repeat("я", 2997) + "..."},
		{"backticks", "```\n*bold* https://amazon.com", "```\n*bold* https://amazon.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks, _, err := block.Build(&modules.Notification{
				User:    &database.User{},
				Payload: &database.Payload{Name: "test"},
				Event: &database.Event{
					Protocol:   "http",
					RemoteAddr: "10.13.37.1:1337",
				},
			}, []string{tt.in})
			require.NoError(t, err)
			require.Equal(t, tt.want, codeBlockText(t, blocks[len(blocks)-1]))
		})
	}
}

func codeBlockText(t *testing.T, b slack.Block) string {
	t.Helper()

	rt, ok := b.(*slack.RichTextBlock)
	require.True(t, ok)
	require.Len(t, rt.Elements, 1)

	pre, ok := rt.Elements[0].(*slack.RichTextPreformatted)
	require.True(t, ok)
	require.Len(t, pre.Elements, 1)

	el, ok := pre.Elements[0].(*slack.RichTextSectionTextElement)
	require.True(t, ok)

	return el.Text
}

func TestBuildMinimal(t *testing.T) {
	receivedAt, _ := time.Parse("2006-01-02T15:04:05Z", "2023-01-01T00:00:00Z")

	blocks, text, err := block.Build(&modules.Notification{
		User:    &database.User{},
		Payload: &database.Payload{Name: "test"},
		Event: &database.Event{
			Protocol:   "dns",
			Data:       [][]byte{[]byte("test")},
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
	require.Equal(t, "[test] :mag: DNS", headerBlock.Text.Text)

	require.Equal(t, "[test] DNS from 10.13.37.1", text)
}

func TestBuildWithEmail(t *testing.T) {
	receivedAt, _ := time.Parse("2006-01-02T15:04:05Z", "2023-01-01T00:00:00Z")

	blocks, _, err := block.Build(&modules.Notification{
		User:    &database.User{},
		Payload: &database.Payload{Name: "test"},
		Event: &database.Event{
			Protocol: "smtp",
			Data:     [][]byte{[]byte("test")},
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

func TestBuildBatch(t *testing.T) {
	blocks, text := block.BuildBatch("test", 3, "DNS 2 · HTTP 1")
	require.Len(t, blocks, 2)

	headerBlock, ok := blocks[0].(*slack.HeaderBlock)
	require.True(t, ok)
	require.Equal(t, "[test] 3 more events", headerBlock.Text.Text)

	summaryBlock, ok := blocks[1].(*slack.SectionBlock)
	require.True(t, ok)
	require.Equal(t, "DNS 2 · HTTP 1", summaryBlock.Text.Text)

	require.Equal(t, "[test] 3 more events", text)
}
