// Package remotesvc implements [service.Service] by talking to an api2 server over
// HTTP. It is the client-side mirror of dbsvc: callers use the same
// service.Service abstraction whether running locally against the database or
// remotely against the API. It depends only on the Go standard library.
package remotesvc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/nt0xa/sonar/internal/service"
)

type Service struct {
	baseURL string
	token   string
	http    *http.Client
}

// New returns a Service that sends requests to baseURL authenticating with
// token. If client is nil, a default *http.Client is used; callers configure
// TLS, proxy and timeouts via their own client.
func New(baseURL, token string, client *http.Client) *Service {
	if client == nil {
		client = &http.Client{}
	}
	return &Service{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    client,
	}
}

var _ service.Service = (*Service)(nil)

// do sends an HTTP request to path (which must already include any query
// string), JSON-encoding body when it is non-nil, and decodes a 2xx response
// into out (when out is non-nil). Any non-2xx response is converted into a
// [service.Error].
func (s *Service) do(ctx context.Context, method, path string, body, out any) error {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, s.baseURL+path, rdr)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+s.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := s.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeError(resp)
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return err
		}
	}

	return nil
}

// errorResponse mirrors api2's error body shape.
type errorResponse struct {
	Message  string            `json:"message"`
	Problems map[string]string `json:"problems,omitempty"`
}

// decodeError reconstructs a [service.Error] from a non-200 api2 response. The
// body is api2's {"message", "problems"} shape; the status code determines the
// error kind.
func decodeError(resp *http.Response) error {
	var body errorResponse
	// Best-effort: a missing or malformed body still yields a kind-only error.
	_ = json.NewDecoder(resp.Body).Decode(&body)

	switch resp.StatusCode {
	case http.StatusBadRequest:
		return service.Error{Kind: service.ErrorKindBadRequest, Message: body.Message}
	case http.StatusUnauthorized:
		return service.Error{Kind: service.ErrorKindUnauthorized, Message: body.Message}
	case http.StatusForbidden:
		return service.Error{Kind: service.ErrorKindForbidden, Message: body.Message}
	case http.StatusNotFound:
		return service.Error{Kind: service.ErrorKindNotFound, Message: body.Message}
	case http.StatusConflict:
		return service.Error{Kind: service.ErrorKindConflict, Message: body.Message}
	case http.StatusUnprocessableEntity:
		return service.Error{Kind: service.ErrorKindValidation, Message: body.Message, Problems: body.Problems}
	default:
		return service.Error{Kind: service.ErrorKindInternal, Message: body.Message}
	}
}

// withQuery appends the encoded query to path when non-empty.
func withQuery(path string, q url.Values) string {
	if e := q.Encode(); e != "" {
		return path + "?" + e
	}
	return path
}
