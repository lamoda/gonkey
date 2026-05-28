package models

import "fmt"

// ErrorCategory defines the type of check that failed
type ErrorCategory string

const (
	ErrorCategoryStatusCode     ErrorCategory = "status_code"
	ErrorCategoryResponseBody   ErrorCategory = "body"
	ErrorCategoryResponseHeader ErrorCategory = "header"
	ErrorCategoryDatabase       ErrorCategory = "database"
	ErrorCategoryMock           ErrorCategory = "mock"
)

// CheckError represents a typed error from a specific check
type CheckError struct {
	Category   ErrorCategory
	Identifier string
	Message    string
	Err        error
}

func (e *CheckError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Err.Error())
	}

	return e.Message
}

func (e *CheckError) Unwrap() error {
	return e.Err
}

func (e *CheckError) GetCategory() ErrorCategory {
	return e.Category
}

func (e *CheckError) GetIdentifier() string {
	return e.Identifier
}

func NewStatusCodeError(expected, actual int) error {
	return &CheckError{
		Category: ErrorCategoryStatusCode,
		Message:  fmt.Sprintf("status code mismatch: expected %d, got %d", expected, actual),
	}
}

func NewBodyError(msg string, args ...interface{}) error {
	return &CheckError{
		Category: ErrorCategoryResponseBody,
		Message:  fmt.Sprintf(msg, args...),
	}
}

func NewBodyErrorWithCause(err error, msg string, args ...interface{}) error {
	return &CheckError{
		Category: ErrorCategoryResponseBody,
		Message:  fmt.Sprintf(msg, args...),
		Err:      err,
	}
}

func NewHeaderError(msg string, args ...interface{}) error {
	return &CheckError{
		Category: ErrorCategoryResponseHeader,
		Message:  fmt.Sprintf(msg, args...),
	}
}

func NewDatabaseError(msg string, args ...interface{}) error {
	return &CheckError{
		Category: ErrorCategoryDatabase,
		Message:  fmt.Sprintf(msg, args...),
	}
}

func NewDatabaseErrorWithCause(err error, msg string, args ...interface{}) error {
	return &CheckError{
		Category: ErrorCategoryDatabase,
		Message:  fmt.Sprintf(msg, args...),
		Err:      err,
	}
}

func NewDatabaseErrorWithIdentifier(queryIndex int, msg string, args ...interface{}) error {
	return &CheckError{
		Category:   ErrorCategoryDatabase,
		Identifier: fmt.Sprintf("%d", queryIndex),
		Message:    fmt.Sprintf(msg, args...),
	}
}

func NewMockError(msg string, args ...interface{}) error {
	return &CheckError{
		Category: ErrorCategoryMock,
		Message:  fmt.Sprintf(msg, args...),
	}
}

func NewMockErrorWithService(serviceName, msg string, args ...interface{}) error {
	return &CheckError{
		Category:   ErrorCategoryMock,
		Identifier: serviceName,
		Message:    fmt.Sprintf(msg, args...),
	}
}
