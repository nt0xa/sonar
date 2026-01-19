package httpx

import "net/http"

// Request represents HTTP request metadata for events.
type Request struct {
	Method  string      `json:"method"`
	Proto   string      `json:"proto"`
	URL     string      `json:"url"`
	Host    string      `json:"host"`
	Headers http.Header `json:"headers"`
	Body    string      `json:"body"`
}

// Response represents HTTP response metadata for events.
type Response struct {
	Status  int         `json:"status"`
	Headers http.Header `json:"headers"`
	Body    string      `json:"body"`
}

// Meta contains HTTP-specific event metadata.
type Meta struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}
