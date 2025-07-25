package server

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/fatih/structs"
	"go.opentelemetry.io/otel/attribute"
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

	counter, err := tel.NewInt64UpDownCounter(
		"http.requests.inflight",
		"{count}",
		"Number of requests currently being processed by the server",
	)
	if err != nil {
		panic(err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		start := time.Now()

		ctx, id := withEventID(ctx)

		attrs := []attribute.KeyValue{
			attribute.String("event.id", id.String()),
		}

		if r.Method != "" {
			attrs = append(attrs, attribute.String("http.method", r.Method))
		} else {
			attrs = append(attrs, attribute.String("http.method", http.MethodGet))
		}

		if r.ContentLength >= 0 {
			attrs = append(attrs, attribute.Int64("http.request.content_length", r.ContentLength))
		}

		ctx, span := tel.TraceStart(ctx, "http",
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attrs...,
			),
		)
		defer span.End()

		counter.Add(ctx, 1)

		next.ServeHTTP(w, r.WithContext(ctx))

		counter.Add(ctx, -1)
		requestDuration.Record(ctx, time.Since(start).Milliseconds())
	})
}

func HTTPHandler(
	db *database.DB,
	tel telemetry.Telemetry,
	origin string,
	notify func(context.Context, *httpx.Event),
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
