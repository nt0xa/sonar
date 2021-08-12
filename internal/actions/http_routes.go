package actions

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/spf13/cobra"

	"github.com/bi-zone/sonar/internal/database/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
	"github.com/bi-zone/sonar/internal/utils/valid"
)

type HTTPActions interface {
	HTTPRoutesCreate(context.Context, HTTPRoutesCreateParams) (HTTPRoutesCreateResult, errors.Error)
	HTTPRoutesDelete(context.Context, HTTPRoutesDeleteParams) (HTTPRoutesDeleteResult, errors.Error)
	HTTPRoutesList(context.Context, HTTPRoutesListParams) (HTTPRoutesListResult, errors.Error)
}

type HTTPRoutesHandler interface {
	HTTPRoutesCreate(context.Context, HTTPRoutesCreateResult)
	HTTPRoutesList(context.Context, HTTPRoutesListResult)
	HTTPRoutesDelete(context.Context, HTTPRoutesDeleteResult)
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

type HTTPRoutesCreateResult *HTTPRoute

func HTTPRoutesCreateCommand(p *HTTPRoutesCreateParams) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "new BODY",
		Short: "Create new HTTP route",
		Args:  oneArg("BODY"),
	}

	var (
		headers []string
	)

	cmd.Flags().StringVarP(&p.PayloadName, "payload", "p", "", "Payload name")
	cmd.Flags().StringVarP(&p.Method, "method", "m", "GET",
		fmt.Sprintf("Request method (one of %s)", quoteAndJoin(models.HTTPMethods)))
	cmd.Flags().StringVarP(&p.Path, "path", "P", "/", "Request path")
	cmd.Flags().StringArrayVarP(&headers, "header", "H", []string{}, "Response header")
	cmd.Flags().IntVarP(&p.Code, "code", "c", 200, "Response status code")
	cmd.Flags().BoolVarP(&p.IsDynamic, "dynamic", "d", false, "Interpret body and headers as templates")

	return cmd, func(cmd *cobra.Command, args []string) errors.Error {
		hh := make(map[string][]string)
		for _, header := range headers {
			if !strings.Contains(header, ":") {
				return errors.Validationf(`header %q must contain ":"`, header)
			}
			parts := strings.SplitN(header, ":", 2)
			name, value := parts[0], strings.TrimLeft(parts[1], " ")

			if h, ok := hh[name]; ok {
				h = append(h, value)
			} else {
				hh[name] = []string{value}
			}
		}
		p.Headers = hh

		body := []byte(args[0])
		p.Body = base64.StdEncoding.EncodeToString(body)

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

type HTTPRoutesDeleteResult *HTTPRoute

func HTTPRoutesDeleteCommand(p *HTTPRoutesDeleteParams) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "del INDEX",
		Short: "Delete HTTP route",
		Long:  "Delete HTTP route identified by INDEX",
		Args:  oneArg("INDEX"),
	}

	cmd.Flags().StringVarP(&p.PayloadName, "payload", "p", "", "Payload name")

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

type HTTPRoutesListResult []*HTTPRoute

func HTTPRoutesListCommand(p *HTTPRoutesListParams) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List HTTP routes",
	}

	cmd.Flags().StringVarP(&p.PayloadName, "payload", "p", "", "Payload name")

	return cmd, nil
}
