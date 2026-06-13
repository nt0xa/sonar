package cmd2

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/service"
	"github.com/nt0xa/sonar/pkg/cmdx"
)

func (c *Command) httpRoutesCreate(cmd *cobra.Command) cmdx.RunFunc {
	var (
		in      service.HTTPRoutesCreateInput
		headers []string
		file    bool
	)

	cmd.Use = "new BODY"
	cmd.Args = cobra.ExactArgs(1)

	in.Method = service.HTTPMethodGET

	cmd.Flags().StringVarP(&in.PayloadName, "payload", "p", "", "Payload name")
	cmd.Flags().VarP(&in.Method, "method", "m",
		fmt.Sprintf("Request method (one of %s)", strings.Join(service.HTTPMethodNames(), ", ")))
	cmd.Flags().StringVarP(&in.Path, "path", "P", "/", "Request path")
	cmd.Flags().StringArrayVarP(&headers, "header", "H", []string{}, "Response header")
	cmd.Flags().IntVarP(&in.Code, "code", "c", 200, "Response status code")
	cmd.Flags().BoolVarP(&in.IsDynamic, "dynamic", "d", false, "Interpret body and headers as templates")
	if c.opts.allowFileAccess {
		cmd.Flags().BoolVarP(&file, "file", "f", false, "Treat BODY as path to file")
	}

	_ = cmd.RegisterFlagCompletionFunc("payload", c.completePayloadName)
	_ = cmd.RegisterFlagCompletionFunc("method", completeOne(service.HTTPMethodNames()))

	return func(cmd *cobra.Command, args []string) error {
		hdrs, err := parseHeaders(headers)
		if err != nil {
			return err
		}
		in.Headers = hdrs

		body, err := readBody(args[0], file)
		if err != nil {
			return err
		}
		in.Body = base64.StdEncoding.EncodeToString(body)

		if err := validate(in); err != nil {
			return err
		}

		out, err := c.svc.HTTPRoutesCreate(cmd.Context(), in)
		if err != nil {
			return err
		}

		return setResult(cmd.Context(), out)
	}
}

func (c *Command) httpRoutesUpdate(cmd *cobra.Command) cmdx.RunFunc {
	var (
		in service.HTTPRoutesUpdateInput

		method    service.HTTPMethod
		path      string
		code      int
		isDynamic bool
		body      string
		headers   []string
		file      bool
	)

	cmd.Use = "mod INDEX"
	cmd.Args = cobra.ExactArgs(1)
	cmd.ValidArgsFunction = c.completeHTTPRoute

	method = service.HTTPMethodGET

	cmd.Flags().StringVarP(&in.Payload, "payload", "p", "", "Payload name")
	cmd.Flags().VarP(&method, "method", "m",
		fmt.Sprintf("Request method (one of %s)", strings.Join(service.HTTPMethodNames(), ", ")))
	cmd.Flags().StringVarP(&path, "path", "P", "/", "Request path")
	cmd.Flags().StringArrayVarP(&headers, "header", "H", []string{}, "Response header")
	cmd.Flags().IntVarP(&code, "code", "c", 200, "Response status code")
	cmd.Flags().BoolVarP(&isDynamic, "dynamic", "d", false, "Interpret body and headers as templates")
	cmd.Flags().StringVarP(&body, "body", "b", "", "Response body")
	if c.opts.allowFileAccess {
		cmd.Flags().BoolVarP(&file, "file", "f", false, "Treat BODY as path to file")
	}

	_ = cmd.RegisterFlagCompletionFunc("payload", c.completePayloadName)
	_ = cmd.RegisterFlagCompletionFunc("method", completeOne(service.HTTPMethodNames()))

	return func(cmd *cobra.Command, args []string) error {
		i, err := parseIndex(args[0])
		if err != nil {
			return err
		}
		in.Index = i

		if cmd.Flags().Changed("method") {
			in.Method = &method
		}
		if cmd.Flags().Changed("path") {
			in.Path = &path
		}
		if cmd.Flags().Changed("header") {
			hdrs, err := parseHeaders(headers)
			if err != nil {
				return err
			}
			in.Headers = hdrs
		}
		if cmd.Flags().Changed("code") {
			in.Code = &code
		}
		if cmd.Flags().Changed("dynamic") {
			in.IsDynamic = &isDynamic
		}
		if cmd.Flags().Changed("body") {
			b, err := readBody(body, file)
			if err != nil {
				return err
			}
			s := base64.StdEncoding.EncodeToString(b)
			in.Body = &s
		}

		if err := validate(in); err != nil {
			return err
		}

		out, err := c.svc.HTTPRoutesUpdate(cmd.Context(), in)
		if err != nil {
			return err
		}

		return setResult(cmd.Context(), out)
	}
}

func (c *Command) httpRoutesDelete(cmd *cobra.Command) cmdx.RunFunc {
	var in service.HTTPRoutesDeleteInput

	cmd.Use = "del INDEX"
	cmd.Args = cobra.ExactArgs(1)
	cmd.ValidArgsFunction = c.completeHTTPRoute

	cmd.Flags().StringVarP(&in.PayloadName, "payload", "p", "", "Payload name")

	_ = cmd.RegisterFlagCompletionFunc("payload", c.completePayloadName)

	return func(cmd *cobra.Command, args []string) error {
		i, err := parseIndex(args[0])
		if err != nil {
			return err
		}
		in.Index = i

		if err := validate(in); err != nil {
			return err
		}

		out, err := c.svc.HTTPRoutesDelete(cmd.Context(), in)
		if err != nil {
			return err
		}

		return setResult(cmd.Context(), out)
	}
}

func (c *Command) httpRoutesList(cmd *cobra.Command) cmdx.RunFunc {
	var in service.HTTPRoutesListInput

	cmd.Use = "list"
	cmd.Args = cobra.NoArgs

	cmd.Flags().StringVarP(&in.PayloadName, "payload", "p", "", "Payload name")

	_ = cmd.RegisterFlagCompletionFunc("payload", c.completePayloadName)

	return func(cmd *cobra.Command, args []string) error {
		if err := validate(in); err != nil {
			return err
		}

		out, err := c.svc.HTTPRoutesList(cmd.Context(), in)
		if err != nil {
			return err
		}

		return setResult(cmd.Context(), out)
	}
}

func (c *Command) httpRoutesClear(cmd *cobra.Command) cmdx.RunFunc {
	var in service.HTTPRoutesClearInput

	cmd.Use = "clr"
	cmd.Args = cobra.NoArgs

	cmd.Flags().StringVarP(&in.PayloadName, "payload", "p", "", "Payload name")
	cmd.Flags().StringVarP(&in.Path, "path", "P", "", "Path")

	_ = cmd.RegisterFlagCompletionFunc("payload", c.completePayloadName)

	return func(cmd *cobra.Command, args []string) error {
		if err := validate(in); err != nil {
			return err
		}

		out, err := c.svc.HTTPRoutesClear(cmd.Context(), in)
		if err != nil {
			return err
		}

		return setResult(cmd.Context(), out)
	}
}
