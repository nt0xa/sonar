package api2

import (
	"net/http"

	"github.com/nt0xa/sonar/internal/service"
)

type HTTPRoutesCreateRequest struct {
	PayloadName string              `json:"payloadName"`
	Method      service.HTTPMethod  `json:"method"`
	Path        string              `json:"path"`
	Code        int                 `json:"code"`
	Headers     map[string][]string `json:"headers"`
	Body        string              `json:"body"`
	IsDynamic   bool                `json:"isDynamic"`
}

func (api *API) HTTPRoutesCreate(w http.ResponseWriter, r *http.Request) {
	var req HTTPRoutesCreateRequest

	if err := api.decodeJSON(r, &req); err != nil {
		api.handleError(w, r, err)
		return
	}

	route, err := api.svc.HTTPRoutesCreate(r.Context(), service.HTTPRoutesCreateInput{
		PayloadName: req.PayloadName,
		Method:      req.Method,
		Path:        req.Path,
		Code:        req.Code,
		Headers:     req.Headers,
		Body:        req.Body,
		IsDynamic:   req.IsDynamic,
	})
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusOK, route)
}
