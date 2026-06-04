package service

import "errors"

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrConflict     = errors.New("conflict")
	ErrNotFound     = errors.New("not found")
	ErrValidation   = errors.New("validation")
)
