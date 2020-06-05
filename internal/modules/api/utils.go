package api

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/mitchellh/mapstructure"
)

func pathDecode(r *http.Request, dst interface{}) error {
	pp := chi.RouteContext(r.Context()).URLParams

	pmap := make(map[string]string)

	for i, name := range pp.Keys {
		pmap[name] = pp.Values[i]
	}

	return mapstructure.Decode(pmap, dst)
}
