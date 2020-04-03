package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func parseJSON(w http.ResponseWriter, r *http.Request, dst interface{}) (error, string) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			return err, fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return err, "Request body contains badly-formed JSON"

		case errors.As(err, &unmarshalTypeError):
			return err, fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return err, fmt.Sprintf("Request body contains unknown field %s", fieldName)

		case errors.Is(err, io.EOF):
			return err, "Request body must not be empty"

		case err.Error() == "http: request body too large":
			return err, "Request body must not be larger than 1MB"

		default:
			return err, "Unknown error"
		}
	}

	if dec.More() {
		return errors.New("multple json objects"), "request body must only contain a single JSON object"
	}

	return nil, ""
}
