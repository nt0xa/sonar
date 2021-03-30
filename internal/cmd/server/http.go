package server

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/httpdb"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils"
	"github.com/bi-zone/sonar/pkg/httpx"
	"github.com/fatih/structs"
)

// TODO: as parameters
const (
	httpHandlerTimeout = time.Second * 10
	httpMaxBodyBytes   = 1 << 20
)

func HTTPDefault(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	rnd, _ := utils.GenerateRandomString(8)
	w.Write([]byte(fmt.Sprintf("<html><body>%s</body></html>", rnd)))
}

func HTTPHandler(db *database.DB, origin string, notify func(*httpx.Event)) http.Handler {
	return http.TimeoutHandler(
		httpx.BodyReaderHandler(
			httpx.MaxBytesHandler(
				httpx.NotifyHandler(
					notify,
					httpdb.Handler(
						&httpdb.Routes{DB: db, Origin: origin},
						http.HandlerFunc(HTTPDefault),
					),
				),
				httpMaxBodyBytes,
			),
			httpMaxBodyBytes,
		),
		httpHandlerTimeout,
		"timeout",
	)
}

func HTTPEvent(e *httpx.Event) *models.Event {

	type Request struct {
		Method  string      `structs:"method"`
		Proto   string      `structs:"proto"`
		URL     string      `structs:"url"`
		Host    string      `structs:"host"`
		Headers http.Header `structs:"headers"`
		Body    string      `structs:"body"`
	}

	type Response struct {
		Status  int         `structs:"status"`
		Headers http.Header `structs:"headers"`
		Body    string      `structs:"body"`
	}

	type Meta struct {
		Request  Request  `structs:"request"`
		Response Response `structs:"response"`
		Secure   bool     `structs:"secure"`
	}

	meta := &Meta{
		Request: Request{
			Method:  e.Request.Method,
			Proto:   e.Request.Proto,
			Headers: e.Request.Header,
			Host:    e.Request.Host,
			URL:     e.Request.URL.String(),
		},
		Response: Response{
			Status:  e.Response.StatusCode,
			Headers: e.Response.Header,
		},
		Secure: e.Secure,
	}

	reqBody, _ := ioutil.ReadAll(e.Request.Body)
	meta.Request.Body = base64.StdEncoding.EncodeToString(reqBody)

	resBody, _ := ioutil.ReadAll(e.Response.Body)
	meta.Response.Body = base64.StdEncoding.EncodeToString(resBody)

	var proto models.Proto

	if e.Secure {
		proto = models.ProtoHTTPS
	} else {
		proto = models.ProtoHTTP
	}

	return &models.Event{
		Protocol:   proto,
		R:          e.RawRequest,
		W:          e.RawResponse,
		RW:         append(e.RawRequest[:], e.RawResponse...),
		Meta:       models.Meta(structs.Map(meta)),
		RemoteAddr: e.RemoteAddr.String(),
		ReceivedAt: e.ReceivedAt,
	}
}
