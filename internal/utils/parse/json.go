package parse

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

func JSON(src io.Reader, dst interface{}) error {
	dec := json.NewDecoder(src)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("badly-formed json (at position %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("badly-formed json")

		case errors.As(err, &unmarshalTypeError):
			return fmt.Errorf("invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)

		case strings.HasPrefix(err.Error(), "unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "unknown field ")
			return fmt.Errorf("unknown field %q", fieldName)

		case errors.Is(err, io.EOF):
			return errors.New("empty")

		case err.Error() == "http: request body too large":
			return errors.New("too large")

		default:
			return errors.New("unknown error")
		}
	}

	if dec.More() {
		return errors.New("multple json objects")
	}

	return nil
}
