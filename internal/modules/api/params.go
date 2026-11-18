package api

import (
	"net/http"
	"strconv"
	"time"
)

func pathInt64(r *http.Request, name string) (int64, error) {
	v, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil {
		return 0, httpError{Status: http.StatusBadRequest, Message: name + " must be an integer"}
	}
	return v, nil
}

func queryUint(r *http.Request, key string) uint {
	v, err := strconv.ParseUint(r.URL.Query().Get(key), 10, 64)
	if err != nil {
		return 0
	}
	return uint(v)
}

func queryInt64Ptr(r *http.Request, key string) (*int64, error) {
	s := r.URL.Query().Get(key)
	if s == "" {
		return nil, nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, httpError{Status: http.StatusBadRequest, Message: key + " must be an integer"}
	}
	return &v, nil
}

func queryTimePtr(r *http.Request, key string) (*time.Time, error) {
	s := r.URL.Query().Get(key)
	if s == "" {
		return nil, nil
	}
	v, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil, httpError{Status: http.StatusBadRequest, Message: key + " must be an RFC3339 timestamp"}
	}
	return &v, nil
}
