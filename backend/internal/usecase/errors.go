package usecase

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("resource conflict")
	ErrBadRequest   = errors.New("bad request")
)

type DomainError struct {
	Err     error
	Message string
}

func (e *DomainError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%v: %s", e.Err, e.Message)
	}
	return e.Err.Error()
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

func NewNotFoundError(msg string) error {
	return &DomainError{Err: ErrNotFound, Message: msg}
}

func NewUnauthorizedError(msg string) error {
	return &DomainError{Err: ErrUnauthorized, Message: msg}
}

func NewForbiddenError(msg string) error {
	return &DomainError{Err: ErrForbidden, Message: msg}
}

func NewConflictError(msg string) error {
	return &DomainError{Err: ErrConflict, Message: msg}
}

func NewBadRequestError(msg string) error {
	return &DomainError{Err: ErrBadRequest, Message: msg}
}
