package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gorilla/schema"
	"github.com/mitchellh/mapstructure"

	"github.com/bi-zone/sonar/internal/utils/errors"
	"github.com/bi-zone/sonar/internal/utils/parse"
)

var decoder = schema.NewDecoder()

func fromPath(r *http.Request, dst interface{}) error {
	pp := chi.RouteContext(r.Context()).URLParams

	pmap := make(map[string]string)

	for i, name := range pp.Keys {
		pmap[name] = pp.Values[i]
	}

	if err := mapstructure.Decode(pmap, dst); err != nil {
		return errors.BadFormatf("path: %s", err)
	}

	return nil
}

func fromQuery(r *http.Request, dst interface{}) error {
	if err := decoder.Decode(dst, r.URL.Query()); err != nil {
		return errors.BadFormatf("query: %s", err)
	}

	return nil
}

func fromJSON(r *http.Request, dst interface{}) error {
	rdr := http.MaxBytesReader(nil, r.Body, 1024*1024)

	if err := parse.JSON(rdr, dst); err != nil {
		return errors.BadFormatf("json: %s", err)
	}

	return nil
}

func responseJSON(w http.ResponseWriter, res interface{}, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(res)
}
