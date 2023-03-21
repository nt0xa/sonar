package main

var cmdCode = `package cmd

import (
	"github.com/spf13/cobra"

	"github.com/russtone/sonar/internal/actions"
	"github.com/russtone/sonar/internal/utils/errors"
)

{{ range . }}
func (c *Command) {{ .Name }}() *cobra.Command {
	{{- if ne .Params.TypeName "" }}
	var params actions.{{ .Params.TypeName }}

	cmd, prepareFunc := actions.{{ .Name }}Command(&params, c.local)
	{{ else }}
	cmd, prepareFunc := actions.{{ .Name }}Command(c.local)
	{{ end }}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				c.handler.OnResult(cmd.Context(), actions.ErrorResult{err})
				return
			}
		}

		{{- if ne .Params.TypeName "" }}
		if err := params.Validate(); err != nil {
			c.handler.OnResult(cmd.Context(), actions.ErrorResult{errors.Validation(err)})
			return
		}
		{{ end }}

		res, err := c.actions.{{ .Name }}(cmd.Context(){{ if ne .Params.TypeName "" }}, params{{ end }})
		if err != nil {
			c.handler.OnResult(cmd.Context(), actions.ErrorResult{err})
			return
		}

		c.handler.OnResult(cmd.Context(), res)

		return
	}


	return cmd
}
{{ end }}
`

var apiCode = `package api

import (
	"net/http"

	"github.com/russtone/sonar/internal/actions"
)

{{ range . }}
func (api *API) {{ .Name }}(w http.ResponseWriter, r *http.Request) {
	{{ if ne .Params.TypeName "" }}
	var params actions.{{ .Params.TypeName }}
	{{ end }}

	{{- range .Params.Types }}
	if err := from{{ . }}(r, &params); err != nil {
		api.handleError(w, r, err)
		return
	}
	{{ end }}

	res, err := api.actions.{{ .Name }}(r.Context(){{ if ne .Params.TypeName "" }}, params{{ end }})
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	responseJSON(w, res, {{ if contains "Create" .Name }}http.StatusCreated{{ else }}http.StatusOK{{ end}})
}
{{ end }}
`

var apiClientCode = `
package apiclient

import (
	"context"

	"github.com/russtone/sonar/internal/actions"
	"github.com/russtone/sonar/internal/utils/errors"
)

{{ range . }}
func (c *Client) {{ .Name }}(ctx context.Context{{ if ne .Params.TypeName "" }}, params actions.{{ .Params.TypeName }}{{ end }}) ({{ .Result }}, errors.Error) {
	var res {{ .Result }}

	err := handle(c.client.R().
		{{- range .Params.Types }}
		{{- if eq . "JSON" }}
		SetBody(params).
		{{- else if eq . "Query" }}
		SetQueryParamsFromValues(toQuery(params)).
		{{- else if eq . "Path" }}
		SetPathParams(toPath(params)).
		{{- end }}
		{{ else }}
		{{- end -}}
		SetError(&APIError{}).
		SetResult(&res).
		SetContext(ctx).
		{{ .HTTPMethod | lower | title }}("{{ .HTTPPath }}"))

	if err != nil {
		return nil, err
	}

	return res, nil
}
{{ end }}
`
