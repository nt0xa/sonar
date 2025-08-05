package card

import (
	"fmt"
	"net"
	"strings"

	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/modules"
	"gopkg.in/square/go-jose.v2/json"
)

type Card struct {
	Schema string `json:"schema"`
	Config Config `json:"config"`
	Body   Body   `json:"body"`
	Header Header `json:"header"`
}

type Config struct {
	UpdateMulti bool     `json:"update_multi"`
	Locales     []string `json:"locales"`
	Style       Style    `json:"style"`
}

type Style struct {
	TextSize TextSizeMap `json:"text_size"`
}

type TextSizeMap map[string]TextSize

type TextSize struct {
	Default string `json:"default"`
	PC      string `json:"pc"`
	Mobile  string `json:"mobile"`
}

type Body struct {
	Direction string    `json:"direction"`
	Padding   string    `json:"padding"`
	Elements  []Element `json:"elements"`
}

type Element struct {
	Tag               string     `json:"tag"`
	HorizontalSpacing string     `json:"horizontal_spacing,omitempty"`
	HorizontalAlign   string     `json:"horizontal_align,omitempty"`
	Columns           []Column   `json:"columns,omitempty"`
	Margin            string     `json:"margin,omitempty"`
	Elements          []Markdown `json:"elements,omitempty"`
	VerticalSpacing   string     `json:"vertical_spacing,omitempty"`
	VerticalAlign     string     `json:"vertical_align,omitempty"`
	Weight            int        `json:"weight,omitempty"`
	Content           string     `json:"content,omitempty"`
	TextAlign         string     `json:"text_align,omitempty"`
	TextSize          string     `json:"text_size,omitempty"`
	Icon              *Icon      `json:"icon,omitempty"`
}

type Column struct {
	Tag             string     `json:"tag"`
	Width           string     `json:"width"`
	Elements        []Markdown `json:"elements"`
	VerticalSpacing string     `json:"vertical_spacing"`
	HorizontalAlign string     `json:"horizontal_align"`
	VerticalAlign   string     `json:"vertical_align"`
	Weight          int        `json:"weight"`
}

type Markdown struct {
	Tag       string `json:"tag"`
	Content   string `json:"content"`
	TextAlign string `json:"text_align"`
	TextSize  string `json:"text_size"`
	Margin    string `json:"margin"`
	Icon      *Icon  `json:"icon,omitempty"`
}

type Icon struct {
	Tag   string `json:"tag"`
	Token string `json:"token"`
	Color string `json:"color,omitempty"`
}

type Header struct {
	Title    HeaderText `json:"title"`
	Subtitle HeaderText `json:"subtitle"`
	Template string     `json:"template"`
	Icon     Icon       `json:"icon"`
	Padding  string     `json:"padding"`
}

type HeaderText struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

func Build(n *modules.Notification, rw []byte) ([]byte, error) {
	host, _, err := net.SplitHostPort(n.Event.RemoteAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to split host port: %w", err)
	}

	body := make([]Element, 0)

	body = append(body, Element{
		Tag:               "column_set",
		HorizontalSpacing: "8px",
		HorizontalAlign:   "left",
		Margin:            "0px 0px 0px 0px",
		Columns: []Column{
			{
				Tag:             "column",
				Width:           "weighted",
				VerticalSpacing: "8px",
				HorizontalAlign: "left",
				VerticalAlign:   "top",
				Weight:          1,
				Elements: []Markdown{
					{
						Tag:       "markdown",
						Content:   fmt.Sprintf("<font color=\"grey\">IP</font>\n%s", host),
						TextAlign: "left",
						TextSize:  "custom",
						Margin:    "0px 0px 0px 0px",
						Icon: &Icon{
							Tag:   "standard_icon",
							Token: "lan_outlined",
							Color: "green",
						},
					},
				},
			},
			{
				Tag:             "column",
				Width:           "weighted",
				VerticalSpacing: "8px",
				HorizontalAlign: "left",
				VerticalAlign:   "top",
				Weight:          1,
				Elements: []Markdown{
					{
						Tag: "markdown",
						Content: fmt.Sprintf(
							"<font color=\"grey\">Time</font>\n%s",
							n.Event.ReceivedAt.Format("02 Jan 2006 15:04:05 MST"),
						),
						TextAlign: "left",
						TextSize:  "custom",
						Margin:    "0px 0px 0px 0px",
						Icon: &Icon{
							Tag:   "standard_icon",
							Token: "time_outlined",
							Color: "blue",
						},
					},
				},
			},
		},
	})

	// TODO: change after Event.Meta is struct
	if geo, ok := n.Event.Meta["geoip"]; ok {
		row := Element{
			Tag:               "column_set",
			HorizontalSpacing: "8px",
			HorizontalAlign:   "left",
			Margin:            "0px 0px 0px 0px",
			Columns: []Column{
				{
					Tag:             "column",
					Width:           "weighted",
					VerticalSpacing: "8px",
					HorizontalAlign: "left",
					VerticalAlign:   "top",
					Weight:          1,
					Elements:        []Markdown{},
				},
				{
					Tag:             "column",
					Width:           "weighted",
					VerticalSpacing: "8px",
					HorizontalAlign: "left",
					VerticalAlign:   "top",
					Weight:          1,
					Elements:        []Markdown{},
				},
			},
		}

		g, _ := geo.(map[string]any)
		if country, ok := g["country"].(map[string]any); ok {
			location := fmt.Sprintf(
				"<font color=\"grey\">Location</font>\n%s %s",
				country["flagEmoji"],
				country["name"],
			)

			if city, ok := g["city"]; ok {
				location += ", " + city.(string)
			}

			row.Columns[0].Elements = []Markdown{{
				Tag:       "markdown",
				Content:   location,
				TextAlign: "left",
				TextSize:  "custom",
				Margin:    "0px 0px 0px 0px",
				Icon: &Icon{
					Tag:   "standard_icon",
					Token: "local_outlined",
					Color: "red",
				},
			}}
		}

		if asn, ok := g["asn"].(map[string]any); ok {
			row.Columns[1].Elements = []Markdown{{
				Tag: "markdown",
				Content: fmt.Sprintf(
					"<font color=\"grey\">Org</font>\n%s (AS%d)",
					asn["org"],
					asn["number"],
				),
				TextAlign: "left",
				TextSize:  "custom",
				Margin:    "0px 0px 0px 0px",
				Icon: &Icon{
					Tag:   "standard_icon",
					Token: "company_outlined",
					Color: "orange",
				},
			}}
		}
		body = append(body, row)
	}

	body = append(body, Element{
		Tag:    "hr",
		Margin: "0px 0px 0px 0px",
	})

	var (
		headerTemplate string
		headerIcon     string
	)

	switch n.Event.Protocol.Category() {
	case models.ProtoCategoryDNS:
		headerTemplate = "carmine"
		headerIcon = "history-search_filled"
	case models.ProtoCategoryFTP:
		headerTemplate = "turquoise"
		headerIcon = "multi-folder_filled"
	case models.ProtoCategorySMTP:
		headerTemplate = "indigo"
		headerIcon = "tab-mail_filled"
	case models.ProtoCategoryHTTP:
		headerTemplate = "wathet"
		headerIcon = "language_filled"
	}

	body = append(body, Element{
		Tag:       "markdown",
		Content:   fmt.Sprintf("```\n%s\n```", rw),
		TextAlign: "left",
		TextSize:  "custom",
		Margin:    "0px 0px 0px 0px",
	})

	card := Card{
		Schema: "2.0",
		Config: Config{
			UpdateMulti: true,
			Locales:     []string{"default"},
			Style: Style{
				TextSize: map[string]TextSize{
					"custom": {
						Default: "normal",
						PC:      "normal",
						Mobile:  "normal",
					},
				},
			},
		},
		Header: Header{
			Title: HeaderText{
				Tag: "plain_text",
				Content: fmt.Sprintf(
					"[%s] %s",
					n.Payload.Name,
					strings.ToUpper(n.Event.Protocol.String()),
				),
			},
			Subtitle: HeaderText{
				Tag:     "plain_text",
				Content: "",
			},
			Template: headerTemplate,
			Icon: Icon{
				Tag:   "standard_icon",
				Token: headerIcon,
			},
			Padding: "12px 12px 12px 12px",
		},
		Body: Body{
			Direction: "vertical",
			Padding:   "12px 12px 12px 12px",
			Elements:  body,
		},
	}

	return json.Marshal(card)
}
