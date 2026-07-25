package domain

import "errors"

type ErrorKind string

const (
	ErrorKindInvalid   ErrorKind = "invalid"
	ErrorKindNotFound  ErrorKind = "not_found"
	ErrorKindConflict  ErrorKind = "conflict"
	ErrorKindInvariant ErrorKind = "invariant"
	ErrorKindInternal  ErrorKind = "internal"
)

type AppError struct {
	Kind    ErrorKind
	Message string
	Err     error
}

func InvalidError(message string) error {
	return AppError{Kind: ErrorKindInvalid, Message: message}
}

func ConflictError(message string) error {
	return AppError{Kind: ErrorKindConflict, Message: message}
}

func InvariantError(message string) error {
	return AppError{Kind: ErrorKindInvariant, Message: message}
}

func IsErrorKind(err error, kind ErrorKind) bool {
	var appErr AppError
	return errors.As(err, &appErr) && appErr.Kind == kind
}

func (e AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return string(e.Kind)
}

func (e AppError) Unwrap() error {
	return e.Err
}
