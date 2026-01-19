// based on https://github.com/DusanKasan/parsemail
package smtpx

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"
	"time"
	"unicode"

	"golang.org/x/net/html"
)

const (
	contentTypeTextPlain = "text/plain"
	contentTypeTextHTML  = "text/html"
)

type Address struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type Email struct {
	Subject string     `json:"subject"`
	From    []Address  `json:"from"`
	To      []Address  `json:"to"`
	Cc      []Address  `json:"cc"`
	Bcc     []Address  `json:"bcc"`
	Date    *time.Time `json:"date"`
	Text    string     `json:"text"`
	HTML    string     `json:"html"`
}

func parseAddressList(s string) []Address {
	res := make([]Address, 0)
	parsed, err := mail.ParseAddressList(s)
	if err != nil {
		return res
	}

	for _, addr := range parsed {
		res = append(res, Address{Name: addr.Name, Address: addr.Address})
	}

	return res
}

// Parse an email message read from io.Reader into parsemail.Email struct, only extracting text/plain.
func Parse(data string) Email {
	var email Email

	msg, err := mail.ReadMessage(strings.NewReader(data))
	if err != nil {
		return email
	}

	email.Subject = decodeMimeSentence(msg.Header.Get("Subject"))
	email.From = parseAddressList(msg.Header.Get("From"))
	email.To = parseAddressList(msg.Header.Get("To"))
	email.Cc = parseAddressList(msg.Header.Get("Cc"))
	email.Bcc = parseAddressList(msg.Header.Get("Bcc"))
	email.Date, _ = parseDate(msg.Header.Get("Date"))

	var (
		text strings.Builder
		html strings.Builder
	)

	parseContent(
		msg.Body,
		&text,
		&html,
		msg.Header.Get("Content-Type"),
		msg.Header.Get("Content-Transfer-Encoding"),
	)

	email.HTML = strings.TrimSpace(html.String())
	email.Text = strings.TrimSpace(text.String())

	if email.Text == "" {
		email.Text = strings.TrimSpace(StripHTML(email.HTML))
	}

	return email
}

func parseContent(
	data io.Reader,
	text StringBuilder,
	html StringBuilder,
	contentType string,
	transferEncoding string,
) {
	ct, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		ct = contentTypeTextPlain
	}

	var builder StringBuilder

	switch ct {
	case contentTypeTextPlain:
		builder = text
	case contentTypeTextHTML:
		builder = html
	}

	if builder != nil {
		if content, err := readContent(
			data,
			transferEncoding,
		); err == nil {
			_, _ = builder.WriteString(content)
		} else {
			return
		}
	} else {
		parseMultipart(
			data,
			text,
			html,
			params["boundary"],
		)
	}
}

func parseMultipart(
	msg io.Reader,
	text StringBuilder,
	html StringBuilder,
	boundary string,
) {
	mr := multipart.NewReader(msg, boundary)
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			return
		}

		parseContent(
			part,
			text,
			html,
			part.Header.Get("Content-Type"),
			part.Header.Get("Content-Transfer-Encoding"),
		)
	}
}

func readContent(r io.Reader, encoding string) (string, error) {
	decoded, err := decodeContent(r, encoding)
	if err != nil {
		return "", err
	}

	b, err := io.ReadAll(decoded)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(b)), nil
}

func decodeContent(content io.Reader, encoding string) (io.Reader, error) {
	encoding = strings.ToLower(strings.TrimSpace(encoding))
	switch encoding {
	case "base64":
		return base64.NewDecoder(base64.StdEncoding, content), nil
	case "quoted-printable":
		return quotedprintable.NewReader(content), nil
	case "7bit", "8bit", "":
		return content, nil
	default:
		return nil, fmt.Errorf("unknown encoding: %s", encoding)
	}
}

func decodeMimeSentence(s string) string {
	result := []string{}
	for word := range strings.SplitSeq(s, " ") {
		dec := new(mime.WordDecoder)
		w, err := dec.Decode(word)
		if err != nil {
			if len(result) == 0 {
				w = word
			} else {
				w = " " + word
			}
		}
		result = append(result, w)
	}
	return strings.Join(result, "")
}

func parseDate(s string) (*time.Time, error) {
	formats := []string{
		time.RFC1123Z,
		time.RFC1123Z + " (MST)",
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"Mon, 2 Jan 2006 15:04:05 -0700 (MST)",
		"Mon, 2 Jan 2006 15:04:05 MST",
	}

	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			tutc := t.UTC()
			return &tutc, nil
		}
	}

	return nil, fmt.Errorf("could not parse date: %s", s)
}

type WhitespaceCollapsingBuilder struct {
	builder       strings.Builder
	prevIsWS      bool
	newLinesCount int
}

func (w *WhitespaceCollapsingBuilder) WriteRune(r rune) (int, error) {
	isWS := unicode.IsSpace(r)
	if isWS && w.prevIsWS && (r != '\n' || (r == '\n' && w.newLinesCount >= 2)) {
		// Skip repeated same whitespace
		return 0, nil
	}
	if isWS {
		if r == '\n' {
			w.newLinesCount += 1
		}
	} else {
		w.newLinesCount = 0
	}
	w.prevIsWS = isWS
	return w.builder.WriteRune(r)
}

func (w *WhitespaceCollapsingBuilder) WriteString(s string) (int, error) {
	n := 0
	for _, r := range s {
		written, err := w.WriteRune(r)
		if err != nil {
			return n, err
		}
		n += written
	}
	return n, nil
}

func (w *WhitespaceCollapsingBuilder) String() string {
	return w.builder.String()
}

func StripHTML(htmlStr string) string {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return htmlStr // fallback: return original if parsing fails
	}

	var buf WhitespaceCollapsingBuilder
	stripHTML(doc, &buf)
	s := buf.String()

	return s
}

type StringBuilder interface {
	WriteString(s string) (int, error)
	WriteRune(r rune) (int, error)
	String() string
}

func stripHTML(n *html.Node, buf StringBuilder) {
	if n.Type == html.ElementNode &&
		(n.Data == "script" || n.Data == "style") {
		return
	}

	if n.Type == html.ElementNode && n.Data == "a" {
		// Collect the inner text of the <a> tag
		var innerText strings.Builder
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			stripHTML(c, &innerText)
		}

		_, _ = buf.WriteString("[")
		_, _ = buf.WriteString(strings.TrimSpace(innerText.String()))
		_, _ = buf.WriteString("]")

		// Find the href attribute
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				_, _ = buf.WriteString("(")
				_, _ = buf.WriteString(attr.Val)
				_, _ = buf.WriteString(")")
				break
			}
		}
		return
	}

	if n.Type == html.TextNode {
		_, _ = buf.WriteString(n.Data)
	}

	// Recurse for all children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		stripHTML(c, buf)
	}

	if n.Type == html.ElementNode &&
		(n.Data == "div" || n.Data == "p" || n.Data == "li" || n.Data == "br") {
		_, _ = buf.WriteString("\n")
	}
}
