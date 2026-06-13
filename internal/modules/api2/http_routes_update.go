package api2

import (
	"net/http"

	"github.com/nt0xa/sonar/internal/modules/api2/apimodels"
	"github.com/nt0xa/sonar/internal/service"
)

func (api *API) HTTPRoutesUpdate(w http.ResponseWriter, r *http.Request) {
	var req apimodels.HTTPRoutesUpdateRequest

	if err := api.decodeJSON(r, &req); err != nil {
		api.handleError(w, r, err)
		return
	}

	index, err := pathInt64(r, "index")
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	route, err := api.svc.HTTPRoutesUpdate(r.Context(), service.HTTPRoutesUpdateInput{
		Payload:   r.PathValue("payload"),
		Index:     index,
		Method:    req.Method,
		Path:      req.Path,
		Code:      req.Code,
		Headers:   req.Headers,
		Body:      req.Body,
		IsDynamic: req.IsDynamic,
	})
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusOK, route)
}
