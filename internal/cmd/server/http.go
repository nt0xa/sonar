package server

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/fatih/structs"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/semconv/v1.13.0/httpconv"
	"go.opentelemetry.io/otel/trace"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/httpdb"
	"github.com/nt0xa/sonar/internal/utils"
	"github.com/nt0xa/sonar/pkg/httpx"
	"github.com/nt0xa/sonar/pkg/telemetry"
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

func HTTPTelemetry(next http.Handler, tel telemetry.Telemetry) http.Handler {
	requestDuration, err := tel.NewInt64Histogram(
		"http.request.duration",
		"ms",
		"HTTP request duration",
	)
	if err != nil {
		panic(err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ctx, span := tel.TraceStart(r.Context(), "http",
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				httpconv.ServerRequest("http", r)...,
			),
		)
		defer span.End()

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)

		requestDuration.Record(
			ctx,
			time.Since(start).Milliseconds(),
			metric.WithAttributes(
				httpconv.ServerRequest("http", r)...,
			),
		)
	})
}

func HTTPHandler(
	db *database.DB,
	tel telemetry.Telemetry,
	origin string,
	notify func(*httpx.Event),
) http.Handler {
	return HTTPTelemetry(
		http.TimeoutHandler(
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
		),
		tel,
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
