package api

import (
	"fmt"
	"net/http"
)

func (api *API) Start() error {
	srv := http.Server{
		Addr:      fmt.Sprintf(":%d", api.cfg.Port),
		Handler:   api.Router(),
		TLSConfig: api.tls,
	}

	return srv.ListenAndServeTLS("", "")
}
