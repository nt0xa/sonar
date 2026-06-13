package cmd2

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/service"
	"github.com/nt0xa/sonar/pkg/cmdx"
)

func (c *Command) addHTTP(g *cmdx.Command) {
	create := &httpRoutesCreate{c: c}
	g.Add("new", "Create new HTTP route", create.run, create.flags)

	update := &httpRoutesUpdate{c: c}
	g.Add("mod", "Update HTTP route", update.run, update.flags)

	del := &httpRoutesDelete{c: c}
	g.Add("del", "Delete HTTP route", del.run, del.flags)

	list := &httpRoutesList{c: c}
	g.Add("list", "List HTTP routes", list.run, list.flags)

	clear := &httpRoutesClear{c: c}
	g.Add("clr", "Delete multiple HTTP routes", clear.run, clear.flags)
}

//
// Create
//

type httpRoutesCreate struct {
	c       *Command
	in      service.HTTPRoutesCreateInput
	headers []string
	file    bool
}

func (x *httpRoutesCreate) flags(cmd *cobra.Command) {
	cmd.Use = "new BODY"
	cmd.Args = cobra.ExactArgs(1)

	x.in.Method = service.HTTPMethodGET

	cmd.Flags().StringVarP(&x.in.PayloadName, "payload", "p", "", "Payload name")
	cmd.Flags().VarP(httpMethodValue{&x.in.Method}, "method", "m",
		fmt.Sprintf("Request method (one of %s)", strings.Join(service.HTTPMethodNames(), ", ")))
	cmd.Flags().StringVarP(&x.in.Path, "path", "P", "/", "Request path")
	cmd.Flags().StringArrayVarP(&x.headers, "header", "H", []string{}, "Response header")
	cmd.Flags().IntVarP(&x.in.Code, "code", "c", 200, "Response status code")
	cmd.Flags().BoolVarP(&x.in.IsDynamic, "dynamic", "d", false, "Interpret body and headers as templates")
	cmd.Flags().BoolVarP(&x.file, "file", "f", false, "Treat BODY as path to file")

	_ = cmd.RegisterFlagCompletionFunc("payload", x.c.completePayloadName)
	_ = cmd.RegisterFlagCompletionFunc("method", completeOne(service.HTTPMethodNames()))
}

func (x *httpRoutesCreate) run(cmd *cobra.Command, args []string) error {
	headers, err := parseHeaders(x.headers)
	if err != nil {
		return err
	}
	x.in.Headers = headers

	body, err := readBody(args[0], x.file)
	if err != nil {
		return err
	}
	x.in.Body = base64.StdEncoding.EncodeToString(body)

	if err := validate(x.in); err != nil {
		return err
	}

	out, err := x.c.svc.HTTPRoutesCreate(cmd.Context(), x.in)
	if err != nil {
		return err
	}

	return setResult(cmd.Context(), out)
}

//
// Update
//

type httpRoutesUpdate struct {
	c  *Command
	in service.HTTPRoutesUpdateInput

	method    service.HTTPMethod
	path      string
	code      int
	isDynamic bool
	body      string
	headers   []string
	file      bool
}

func (x *httpRoutesUpdate) flags(cmd *cobra.Command) {
	cmd.Use = "mod INDEX"
	cmd.Args = cobra.ExactArgs(1)
	cmd.ValidArgsFunction = x.c.completeHTTPRoute

	x.method = service.HTTPMethodGET

	cmd.Flags().StringVarP(&x.in.Payload, "payload", "p", "", "Payload name")
	cmd.Flags().VarP(httpMethodValue{&x.method}, "method", "m",
		fmt.Sprintf("Request method (one of %s)", strings.Join(service.HTTPMethodNames(), ", ")))
	cmd.Flags().StringVarP(&x.path, "path", "P", "/", "Request path")
	cmd.Flags().StringArrayVarP(&x.headers, "header", "H", []string{}, "Response header")
	cmd.Flags().IntVarP(&x.code, "code", "c", 200, "Response status code")
	cmd.Flags().BoolVarP(&x.isDynamic, "dynamic", "d", false, "Interpret body and headers as templates")
	cmd.Flags().StringVarP(&x.body, "body", "b", "", "Response body")
	cmd.Flags().BoolVarP(&x.file, "file", "f", false, "Treat BODY as path to file")

	_ = cmd.RegisterFlagCompletionFunc("payload", x.c.completePayloadName)
	_ = cmd.RegisterFlagCompletionFunc("method", completeOne(service.HTTPMethodNames()))
}

func (x *httpRoutesUpdate) run(cmd *cobra.Command, args []string) error {
	i, err := parseIndex(args[0])
	if err != nil {
		return err
	}
	x.in.Index = i

	if cmd.Flags().Changed("method") {
		x.in.Method = &x.method
	}
	if cmd.Flags().Changed("path") {
		x.in.Path = &x.path
	}
	if cmd.Flags().Changed("header") {
		headers, err := parseHeaders(x.headers)
		if err != nil {
			return err
		}
		x.in.Headers = headers
	}
	if cmd.Flags().Changed("code") {
		x.in.Code = &x.code
	}
	if cmd.Flags().Changed("dynamic") {
		x.in.IsDynamic = &x.isDynamic
	}
	if cmd.Flags().Changed("body") {
		body, err := readBody(x.body, x.file)
		if err != nil {
			return err
		}
		s := base64.StdEncoding.EncodeToString(body)
		x.in.Body = &s
	}

	if err := validate(x.in); err != nil {
		return err
	}

	out, err := x.c.svc.HTTPRoutesUpdate(cmd.Context(), x.in)
	if err != nil {
		return err
	}

	return setResult(cmd.Context(), out)
}

//
// Delete
//

type httpRoutesDelete struct {
	c  *Command
	in service.HTTPRoutesDeleteInput
}

func (x *httpRoutesDelete) flags(cmd *cobra.Command) {
	cmd.Use = "del INDEX"
	cmd.Args = cobra.ExactArgs(1)
	cmd.ValidArgsFunction = x.c.completeHTTPRoute

	cmd.Flags().StringVarP(&x.in.PayloadName, "payload", "p", "", "Payload name")

	_ = cmd.RegisterFlagCompletionFunc("payload", x.c.completePayloadName)
}

func (x *httpRoutesDelete) run(cmd *cobra.Command, args []string) error {
	i, err := parseIndex(args[0])
	if err != nil {
		return err
	}
	x.in.Index = i

	if err := validate(x.in); err != nil {
		return err
	}

	out, err := x.c.svc.HTTPRoutesDelete(cmd.Context(), x.in)
	if err != nil {
		return err
	}

	return setResult(cmd.Context(), out)
}

//
// List
//

type httpRoutesList struct {
	c  *Command
	in service.HTTPRoutesListInput
}

func (x *httpRoutesList) flags(cmd *cobra.Command) {
	cmd.Use = "list"
	cmd.Args = cobra.NoArgs

	cmd.Flags().StringVarP(&x.in.PayloadName, "payload", "p", "", "Payload name")

	_ = cmd.RegisterFlagCompletionFunc("payload", x.c.completePayloadName)
}

func (x *httpRoutesList) run(cmd *cobra.Command, args []string) error {
	if err := validate(x.in); err != nil {
		return err
	}

	out, err := x.c.svc.HTTPRoutesList(cmd.Context(), x.in)
	if err != nil {
		return err
	}

	return setResult(cmd.Context(), out)
}

//
// Clear
//

type httpRoutesClear struct {
	c  *Command
	in service.HTTPRoutesClearInput
}

func (x *httpRoutesClear) flags(cmd *cobra.Command) {
	cmd.Use = "clr"
	cmd.Args = cobra.NoArgs

	cmd.Flags().StringVarP(&x.in.PayloadName, "payload", "p", "", "Payload name")
	cmd.Flags().StringVarP(&x.in.Path, "path", "P", "", "Path")

	_ = cmd.RegisterFlagCompletionFunc("payload", x.c.completePayloadName)
}

func (x *httpRoutesClear) run(cmd *cobra.Command, args []string) error {
	if err := validate(x.in); err != nil {
		return err
	}

	out, err := x.c.svc.HTTPRoutesClear(cmd.Context(), x.in)
	if err != nil {
		return err
	}

	return setResult(cmd.Context(), out)
}
