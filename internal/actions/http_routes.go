package actions

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/utils/errors"
	"github.com/nt0xa/sonar/internal/utils/valid"
)

const (
	HTTPRoutesCreateResultID = "http-routes/create"
	HTTPRoutesUpdateResultID = "http-routes/update"
	HTTPRoutesDeleteResultID = "http-routes/delete"
	HTTPRoutesClearResultID  = "http-routes/clear"
	HTTPRoutesListResultID   = "http-routes/list"
)

type HTTPActions interface {
	HTTPRoutesCreate(context.Context, HTTPRoutesCreateParams) (*HTTPRoutesCreateResult, errors.Error)
	HTTPRoutesUpdate(context.Context, HTTPRoutesUpdateParams) (*HTTPRoutesUpdateResult, errors.Error)
	HTTPRoutesDelete(context.Context, HTTPRoutesDeleteParams) (*HTTPRoutesDeleteResult, errors.Error)
	HTTPRoutesClear(context.Context, HTTPRoutesClearParams) (HTTPRoutesClearResult, errors.Error)
	HTTPRoutesList(context.Context, HTTPRoutesListParams) (HTTPRoutesListResult, errors.Error)
}

type HTTPRoute struct {
	Index            int64               `json:"index"`
	PayloadSubdomain string              `json:"payloadSubdomain"`
	Method           string              `json:"method"`
	Path             string              `json:"path"`
	Code             int                 `json:"code"`
	Headers          map[string][]string `json:"headers"`
	Body             string              `json:"body"`
	IsDynamic        bool                `json:"isDynamic"`
	CreatedAt        time.Time           `json:"createdAt"`
}

//
// Create
//

type HTTPRoutesCreateParams struct {
	PayloadName string              `err:"payloadName" json:"payloadName"`
	Method      string              `err:"method"      json:"method"`
	Path        string              `err:"path"        json:"path"`
	Code        int                 `err:"code"        json:"code"`
	Headers     map[string][]string `err:"headers"     json:"headers"`
	Body        string              `err:"body"        json:"body"`
	IsDynamic   bool                `err:"isDynamic"   json:"isDynamic"`
}

func (p HTTPRoutesCreateParams) Validate() error {
	methods := []string{models.HTTPMethodAny}
	methods = append(methods, models.HTTPMethods...)

	return validation.ValidateStruct(&p,
		validation.Field(&p.PayloadName, validation.Required),
		validation.Field(&p.Method, validation.Required, valid.OneOf(methods, false)),
		validation.Field(&p.Path, validation.Required,
			validation.Match(regexp.MustCompile("^/.*")).Error(`path must start with "/"`)),
		validation.Field(&p.Code, validation.Required),
	)
}

type HTTPRoutesCreateResult struct {
	HTTPRoute
}

func (r HTTPRoutesCreateResult) ResultID() string {
	return HTTPRoutesCreateResultID
}

func HTTPRoutesCreateCommand(acts *Actions, p *HTTPRoutesCreateParams, local bool) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "new BODY",
		Short: "Create new HTTP route",
		Args:  oneArg("BODY"),
	}

	var (
		headers []string
		file    bool
	)

	methods := append([]string{models.HTTPMethodAny}, models.HTTPMethods...)

	cmd.Flags().StringVarP(&p.PayloadName, "payload", "p", "", "Payload name")
	cmd.Flags().StringVarP(&p.Method, "method", "m", "GET",
		fmt.Sprintf("Request method (one of %s)", quoteAndJoin(methods)))
	cmd.Flags().StringVarP(&p.Path, "path", "P", "/", "Request path")
	cmd.Flags().StringArrayVarP(&headers, "header", "H", []string{}, "Response header")
	cmd.Flags().IntVarP(&p.Code, "code", "c", 200, "Response status code")
	cmd.Flags().BoolVarP(&p.IsDynamic, "dynamic", "d", false, "Interpret body and headers as templates")

	// Add file flag only for local client, i.e. terminal.
	// Otherwise anyone will be able to read files from server using telegram client.
	if local {
		cmd.Flags().BoolVarP(&file, "file", "f", false, "Treat BODY as path to file")
	}

	_ = cmd.RegisterFlagCompletionFunc("payload", completePayloadName(acts))
	_ = cmd.RegisterFlagCompletionFunc("method", completeOne(methods))

	return cmd, func(cmd *cobra.Command, args []string) errors.Error {
		hh := make(map[string][]string)
		for _, header := range headers {
			if !strings.Contains(header, ":") {
				return errors.Validationf(`header %q must contain ":"`, header)
			}
			parts := strings.SplitN(header, ":", 2)
			name, value := parts[0], strings.TrimLeft(parts[1], " ")

			hh[name] = append(hh[name], value)
		}
		p.Headers = hh

		var body []byte

		if file {
			b, err := os.ReadFile(args[0])
			if err != nil {
				return errors.Validationf("fail to read file %q", args[0])
			}
			body = b
		} else {
			body = []byte(args[0])
		}

		p.Body = base64.StdEncoding.EncodeToString(body)

		return nil
	}
}

//
// Update
//

type HTTPRoutesUpdateParams struct {
	Payload   string              `err:"payload"     path:"payload"          json:"-"`
	Index     int64               `err:"index"       path:"index"            json:"-"`
	Method    *string             `err:"method"      json:"method,omitempty"`
	Path      *string             `err:"path"        json:"path,omitempty"`
	Code      *int                `err:"code"        json:"code,omitempty"`
	Headers   map[string][]string `err:"headers"     json:"headers,omitempty"`
	Body      *string             `err:"body"        json:"body,omitempty"`
	IsDynamic *bool               `err:"isDynamic"   json:"isDynamic,omitempty"`
}

func (p HTTPRoutesUpdateParams) Validate() error {
	methods := []string{models.HTTPMethodAny}
	methods = append(methods, models.HTTPMethods...)

	return validation.ValidateStruct(&p,
		validation.Field(&p.Payload, validation.Required),
		validation.Field(&p.Method,
			validation.When(p.Method != nil, valid.OneOf(methods, false))),
		validation.Field(&p.Path,
			validation.When(p.Path != nil, validation.Match(regexp.MustCompile("^/.*")).Error(`path must start with "/"`))),
	)
}

type HTTPRoutesUpdateResult struct {
	HTTPRoute
}

func (r HTTPRoutesUpdateResult) ResultID() string {
	return HTTPRoutesUpdateResultID
}

func HTTPRoutesUpdateCommand(acts *Actions, p *HTTPRoutesUpdateParams, local bool) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:               "mod INDEX",
		Short:             "Update HTTP route",
		Args:              oneArg("INDEX"),
		ValidArgsFunction: completeHTTPRoute(acts),
	}

	var (
		headers []string
		file    bool

		method    string
		path      string
		code      int
		isDynamic bool
		body      string
	)

	methods := append([]string{models.HTTPMethodAny}, models.HTTPMethods...)

	cmd.Flags().StringVarP(&p.Payload, "payload", "p", "", "Payload name")
	cmd.Flags().StringVarP(&method, "method", "m", "GET",
		fmt.Sprintf("Request method (one of %s)", quoteAndJoin(methods)))
	cmd.Flags().StringVarP(&path, "path", "P", "/", "Request path")
	cmd.Flags().StringArrayVarP(&headers, "header", "H", []string{}, "Response header")
	cmd.Flags().IntVarP(&code, "code", "c", 200, "Response status code")
	cmd.Flags().BoolVarP(&isDynamic, "dynamic", "d", false, "Interpret body and headers as templates")
	cmd.Flags().StringVarP(&body, "body", "b", "", "Response body")

	// Add file flag only for local client, i.e. terminal.
	// Otherwise anyone will be able to read files from server using telegram client.
	if local {
		cmd.Flags().BoolVarP(&file, "file", "f", false, "Treat BODY as path to file")
	}

	_ = cmd.RegisterFlagCompletionFunc("payload", completePayloadName(acts))
	_ = cmd.RegisterFlagCompletionFunc("method", completeOne(methods))

	return cmd, func(cmd *cobra.Command, args []string) errors.Error {
		// Index
		i, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return errors.Validationf("invalid integer value %q", args[0])
		}
		p.Index = i

		// Method
		if cmd.Flags().Changed("method") {
			p.Method = &method
		}

		// Path
		if cmd.Flags().Changed("path") {
			p.Path = &path
		}

		// Headers
		if cmd.Flags().Changed("header") {
			hh := make(map[string][]string)
			for _, header := range headers {
				if !strings.Contains(header, ":") {
					return errors.Validationf(`header %q must contain ":"`, header)
				}
				parts := strings.SplitN(header, ":", 2)
				name, value := parts[0], strings.TrimLeft(parts[1], " ")

				hh[name] = append(hh[name], value)
			}
			p.Headers = hh
		}

		// Code
		if cmd.Flags().Changed("code") {
			p.Code = &code
		}

		// IsDynamic
		if cmd.Flags().Changed("dynamic") {
			p.IsDynamic = &isDynamic
		}

		// Body
		if cmd.Flags().Changed("body") {
			var bodyBytes []byte

			if file {
				b, err := os.ReadFile(body)
				if err != nil {
					return errors.Validationf("fail to read file %q", body)
				}
				bodyBytes = b
			} else {
				bodyBytes = []byte(body)
			}
			bodyBase64 := base64.StdEncoding.EncodeToString(bodyBytes)

			p.Body = &bodyBase64
		}

		return nil
	}
}

//
// Delete
//

type HTTPRoutesDeleteParams struct {
	PayloadName string `err:"payload" path:"payload"`
	Index       int64  `err:"index"   path:"index"`
}

func (p HTTPRoutesDeleteParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.PayloadName, validation.Required),
		validation.Field(&p.Index, validation.Required),
	)
}

type HTTPRoutesDeleteResult struct {
	HTTPRoute
}

func (r HTTPRoutesDeleteResult) ResultID() string {
	return HTTPRoutesDeleteResultID
}

func HTTPRoutesDeleteCommand(acts *Actions, p *HTTPRoutesDeleteParams, local bool) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:               "del INDEX",
		Short:             "Delete HTTP route",
		Long:              "Delete HTTP route identified by INDEX",
		Args:              oneArg("INDEX"),
		ValidArgsFunction: completeHTTPRoute(acts),
	}

	cmd.Flags().StringVarP(&p.PayloadName, "payload", "p", "", "Payload name")

	_ = cmd.RegisterFlagCompletionFunc("payload", completePayloadName(acts))

	return cmd, func(cmd *cobra.Command, args []string) errors.Error {
		i, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return errors.Validationf("invalid integer value %q", args[0])
		}
		p.Index = i
		return nil
	}
}

//
// Clear
//

type HTTPRoutesClearParams struct {
	PayloadName string `err:"payload" path:"payload"`
	Path        string `err:"path"    query:"path"`
}

func (p HTTPRoutesClearParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.PayloadName, validation.Required),
	)
}

type HTTPRoutesClearResult []HTTPRoute

func (r HTTPRoutesClearResult) ResultID() string {
	return HTTPRoutesClearResultID
}

func HTTPRoutesClearCommand(acts *Actions, p *HTTPRoutesClearParams, local bool) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "clr",
		Short: "Delete multiple HTTP routes",
	}

	cmd.Flags().StringVarP(&p.PayloadName, "payload", "p", "", "Payload name")
	cmd.Flags().StringVarP(&p.Path, "path", "P", "", "Path")

	_ = cmd.RegisterFlagCompletionFunc("payload", completePayloadName(acts))

	return cmd, nil
}

//
// List
//

type HTTPRoutesListParams struct {
	PayloadName string `err:"payload" path:"payload"`
}

func (p HTTPRoutesListParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.PayloadName, validation.Required),
	)
}

type HTTPRoutesListResult []HTTPRoute

func (r HTTPRoutesListResult) ResultID() string {
	return HTTPRoutesListResultID
}

func HTTPRoutesListCommand(acts *Actions, p *HTTPRoutesListParams, local bool) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List HTTP routes",
	}

	cmd.Flags().StringVarP(&p.PayloadName, "payload", "p", "", "Payload name")

	_ = cmd.RegisterFlagCompletionFunc("payload", completePayloadName(acts))

	return cmd, nil
}
