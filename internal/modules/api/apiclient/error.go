package apiclient

import (
	"fmt"
	"strings"
)

type APIError struct {
	Msg  string                 `json:"message"`
	Det  string                 `json:"details"`
	Errs map[string]interface{} `json:"errors"`
}

func (e *APIError) Message() string {
	return e.Msg
}

func (e *APIError) Details() string {
	return e.Det
}

func (e *APIError) Error() string {
	det := ""

	if len(e.Errs) > 0 {
		for name, err := range e.Errs {
			switch ee := err.(type) {
			case string:
				det += fmt.Sprintf("%q: %s; ", name, err)
			case map[string]interface{}:
				for i, err := range ee {
					det += fmt.Sprintf(`"%s.%s": %s; `, name, i, err)
				}
			}
		}
	} else {
		det = e.Det
	}

	return fmt.Sprintf("%s: %s", e.Msg, strings.TrimRight(det, " ;"))
}
