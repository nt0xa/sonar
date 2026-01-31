package card_test

import (
	"testing"
	"time"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/modules"
	"github.com/nt0xa/sonar/internal/modules/lark/card/v2"
	"github.com/nt0xa/sonar/pkg/geoipx"
	"github.com/stretchr/testify/require"
	"gopkg.in/square/go-jose.v2/json"
)

var expectedJSON = `{
    "schema": "2.0",
    "config": {
        "update_multi": true,
        "locales": [
            "default"
        ],
        "style": {
            "text_size": {
                "custom": {
                    "default": "normal",
                    "pc": "normal",
                    "mobile": "normal"
                }
            }
        }
    },
    "body": {
        "direction": "vertical",
        "padding": "12px 12px 12px 12px",
        "elements": [
            {
                "tag": "column_set",
                "horizontal_spacing": "8px",
                "horizontal_align": "left",
                "columns": [
                    {
                        "tag": "column",
                        "width": "weighted",
                        "elements": [
                            {
                                "tag": "markdown",
                                "content": "<font color=\"grey\">IP</font>\n10.13.37.1",
                                "text_align": "left",
                                "text_size": "custom",
                                "margin": "0px 0px 0px 0px",
                                "icon": {
                                    "tag": "standard_icon",
                                    "token": "lan_outlined",
                                    "color": "green"
                                }
                            }
                        ],
                        "vertical_spacing": "8px",
                        "horizontal_align": "left",
                        "vertical_align": "top",
                        "weight": 1
                    },
                    {
                        "tag": "column",
                        "width": "weighted",
                        "elements": [
                            {
                                "tag": "markdown",
                                "content": "<font color=\"grey\">Time</font>\n01 Jan 2023 00:00:00 UTC",
                                "text_align": "left",
                                "text_size": "custom",
                                "margin": "0px 0px 0px 0px",
                                "icon": {
                                    "tag": "standard_icon",
                                    "token": "time_outlined",
                                    "color": "blue"
                                }
                            }
                        ],
                        "vertical_spacing": "8px",
                        "horizontal_align": "left",
                        "vertical_align": "top",
                        "weight": 1
                    }
                ],
                "margin": "0px 0px 0px 0px"
            },
            {
                "tag": "column_set",
                "horizontal_spacing": "8px",
                "horizontal_align": "left",
                "columns": [
                    {
                        "tag": "column",
                        "width": "weighted",
                        "elements": [
                            {
                                "tag": "markdown",
                                "content": "<font color=\"grey\">Location</font>\n🇬🇧 United Kingdom, London",
                                "text_align": "left",
                                "text_size": "custom",
                                "margin": "0px 0px 0px 0px",
                                "icon": {
                                    "tag": "standard_icon",
                                    "token": "local_outlined",
                                    "color": "red"
                                }
                            }
                        ],
                        "vertical_spacing": "8px",
                        "horizontal_align": "left",
                        "vertical_align": "top",
                        "weight": 1
                    },
                    {
                        "tag": "column",
                        "width": "weighted",
                        "elements": [
                            {
                                "tag": "markdown",
                                "content": "<font color=\"grey\">Org</font>\nGoogle Inc. (AS1234)",
                                "text_align": "left",
                                "text_size": "custom",
                                "margin": "0px 0px 0px 0px",
                                "icon": {
                                    "tag": "standard_icon",
                                    "token": "company_outlined",
                                    "color": "orange"
                                }
                            }
                        ],
                        "vertical_spacing": "8px",
                        "horizontal_align": "left",
                        "vertical_align": "top",
                        "weight": 1
                    }
                ],
                "margin": "0px 0px 0px 0px"
            },
            {
                "tag": "hr",
                "margin": "0px 0px 0px 0px"
            },
            {
                "tag": "markdown",
                "content": "` + "```\\ntest\\n```" + `",
                "text_align": "left",
                "text_size": "custom",
                "margin": "0px 0px 0px 0px"
            }
        ]
    },
    "header": {
        "title": {
            "tag": "plain_text",
            "content": "[test] HTTP"
        },
        "subtitle": {
            "tag": "plain_text",
            "content": ""
        },
        "template": "wathet",
        "icon": {
            "tag": "standard_icon",
            "token": "language_filled"
        },
        "padding": "12px 12px 12px 12px"
    }
}`

func TestCard(t *testing.T) {
	receivedAt, _ := time.Parse("2006-01-02T15:04:05Z", "2023-01-01T00:00:00Z")

	card, err := card.Build(&modules.Notification{
		User:    &database.UsersFull{},
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
	}, []byte("test"))

	require.NoError(t, err)

	var expected, got any

	require.NoError(t, json.Unmarshal([]byte(expectedJSON), &expected))
	require.NoError(t, json.Unmarshal(card, &got))

	require.Equal(t, expected, got)
}
