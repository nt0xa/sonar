package service

import (
	"fmt"
	"sort"
	"strings"
)

type ErrorKind int

const (
	ErrorKindInternal ErrorKind = iota
	ErrorKindBadRequest
	ErrorKindUnauthorized
	ErrorKindForbidden
	ErrorKindConflict
	ErrorKindNotFound
	ErrorKindValidation
)

type Error struct {
	Kind     ErrorKind
	Message  string
	Problems map[string]string // field -> problem; validation only
}

func (e Error) Error() string {
	if len(e.Problems) == 0 {
		return e.Message
	}

	// Validation errors carry field-level problems and often no Message; render
	// them deterministically so the message is never empty.
	keys := make([]string, 0, len(e.Problems))
	for k := range e.Problems {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s: %s", k, e.Problems[k]))
	}
	problems := strings.Join(parts, "; ")

	if e.Message != "" {
		return e.Message + ": " + problems
	}
	return problems
}

func BadRequestf(format string, a ...any) Error {
	return Error{Kind: ErrorKindBadRequest, Message: fmt.Sprintf(format, a...)}
}

func Unauthorized() Error { return Error{Kind: ErrorKindUnauthorized, Message: "unauthorized"} }

func Forbiddenf(format string, a ...any) Error {
	return Error{Kind: ErrorKindForbidden, Message: fmt.Sprintf(format, a...)}
}

func NotFoundf(format string, a ...any) Error {
	return Error{Kind: ErrorKindNotFound, Message: fmt.Sprintf(format, a...)}
}

func Conflictf(format string, a ...any) Error {
	return Error{Kind: ErrorKindConflict, Message: fmt.Sprintf(format, a...)}
}

func Validation(problems map[string]string) Error {
	return Error{Kind: ErrorKindValidation, Problems: problems}
}
