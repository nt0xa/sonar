package apiclient

import (
	"fmt"
)

type apiError struct {
	Msg  string            `json:"message"`
	Det  string            `json:"details"`
	Errs map[string]string `json:"errors"`
}

func (e *apiError) Message() string {
	return e.Msg
}

func (e *apiError) Details() string {
	return e.Det
}

func (e *apiError) Error() string {
	det := ""

	if len(e.Errs) > 0 {
		for name, err := range e.Errs {
			det += fmt.Sprintf("%q: %s;", name, err)
		}
	} else {
		det = e.Det
	}

	return fmt.Sprintf("%s: %s", e.Msg, det)
}
