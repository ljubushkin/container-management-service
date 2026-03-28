package apperror

import "errors"

type Code string

const (
	CodeInvalidType       Code = "INVALID_TYPE"
	CodeInvalidStatus     Code = "INVALID_STATUS"
	CodeInvalidWarehouse  Code = "INVALID_WAREHOUSE"
	CodeNotFound          Code = "NOT_FOUND"
	CodeInternal          Code = "INTERNAL_ERROR"
	CodeInvalidPagination Code = "INVALID_PANGINATION"
)

type AppError struct {
	Code    Code
	Message string
	Err     error
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code Code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func Is(err error, code Code) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}
