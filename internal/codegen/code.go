package main

var cmdCode = `package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

{{ range . }}
func (c *command) {{ .Name }}(local bool) *cobra.Command {
	{{- if ne .Params.Name "" }}
	var params actions.{{ .Params.Name }}

	cmd, prepareFunc := actions.{{ .Name }}Command(&params, local)
	{{ else }}
	cmd, prepareFunc := actions.{{ .Name }}Command(local)
	{{ end }}

	cmd.RunE = RunE(func(cmd *cobra.Command, args []string) errors.Error {

		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}

		{{- if ne .Params.Name "" }}
		if err := params.Validate(); err != nil {
			return errors.Validation(err)
		}
		{{ end }}

		res, err := c.actions.{{ .Name }}(cmd.Context(){{ if ne .Params.Name "" }}, params{{ end }})
		if err != nil {
			return err
		}

		c.handler.{{ .Name }}(cmd.Context(), res)

		return nil
	})


	return cmd
}
{{ end }}
`

var apiCode = `package api

import (
	"net/http"

	"github.com/bi-zone/sonar/internal/actions"
)

{{ range . }}
func (api *API) {{ .Name }}(w http.ResponseWriter, r *http.Request) {
	{{ if ne .Params.Name "" }}
	var params actions.{{ .Params.Name }}
	{{ end }}

	{{- range .Params.Types }}
	if err := from{{ . }}(r, &params); err != nil {
		api.handleError(w, r, err)
		return
	}
	{{ end }}

	res, err := api.actions.{{ .Name }}(r.Context(){{ if ne .Params.Name "" }}, params{{ end }})
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

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

{{ range . }}
func (c *Client) {{ .Name }}(ctx context.Context{{ if ne .Params.Name "" }}, params actions.{{ .Params.Name }}{{ end }}) (actions.{{ .Result }}, errors.Error) {
	var res actions.{{ .Result }}

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
