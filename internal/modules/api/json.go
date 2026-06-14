package api

import (
	"encoding/json"
	"net/http"
)

func (api *API) decodeJSON(r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return httpError{Status: http.StatusBadRequest, Message: err.Error()}
	}
	return nil
}

func (api *API) encodeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		api.log.Error("failed to encode response", "error", err)
	}
}
