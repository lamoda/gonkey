package models

import (
	"errors"
	"testing"
)

func TestCheckError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *CheckError
		expected string
	}{
		{
			name: "simple error without cause",
			err: &CheckError{
				Category: ErrorCategoryResponseBody,
				Message:  "body mismatch",
			},
			expected: "body mismatch",
		},
		{
			name: "error with cause",
			err: &CheckError{
				Category: ErrorCategoryDatabase,
				Message:  "query failed",
				Err:      errors.New("connection timeout"),
			},
			expected: "query failed: connection timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCheckError_Unwrap(t *testing.T) {
	cause := errors.New("original error")
	err := &CheckError{
		Category: ErrorCategoryMock,
		Message:  "mock error",
		Err:      cause,
	}

	if unwrapped := err.Unwrap(); unwrapped != cause {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestCheckError_GetCategory(t *testing.T) {
	err := &CheckError{
		Category: ErrorCategoryResponseHeader,
		Message:  "header error",
	}

	if got := err.GetCategory(); got != ErrorCategoryResponseHeader {
		t.Errorf("GetCategory() = %v, want %v", got, ErrorCategoryResponseHeader)
	}
}

func TestNewStatusCodeError(t *testing.T) {
	err := NewStatusCodeError(200, 404)

	var checkErr *CheckError
	if !errors.As(err, &checkErr) {
		t.Fatal("expected CheckError type")
	}

	if checkErr.Category != ErrorCategoryStatusCode {
		t.Errorf("Category = %v, want %v", checkErr.Category, ErrorCategoryStatusCode)
	}

	expectedMsg := "status code mismatch: expected 200, got 404"
	if checkErr.Message != expectedMsg {
		t.Errorf("Message = %v, want %v", checkErr.Message, expectedMsg)
	}
}

func TestNewBodyError(t *testing.T) {
	err := NewBodyError("at path $.id: expected %d, got %d", 123, 456)

	var checkErr *CheckError
	if !errors.As(err, &checkErr) {
		t.Fatal("expected CheckError type")
	}

	if checkErr.Category != ErrorCategoryResponseBody {
		t.Errorf("Category = %v, want %v", checkErr.Category, ErrorCategoryResponseBody)
	}

	expectedMsg := "at path $.id: expected 123, got 456"
	if checkErr.Message != expectedMsg {
		t.Errorf("Message = %v, want %v", checkErr.Message, expectedMsg)
	}
}

func TestNewBodyErrorWithCause(t *testing.T) {
	cause := errors.New("parse error")
	err := NewBodyErrorWithCause(cause, "failed to parse JSON")

	var checkErr *CheckError
	if !errors.As(err, &checkErr) {
		t.Fatal("expected CheckError type")
	}

	if checkErr.Category != ErrorCategoryResponseBody {
		t.Errorf("Category = %v, want %v", checkErr.Category, ErrorCategoryResponseBody)
	}

	if !errors.Is(err, cause) {
		t.Error("expected error to wrap cause")
	}
}

func TestNewHeaderError(t *testing.T) {
	err := NewHeaderError("missing header: %s", "Content-Type")

	var checkErr *CheckError
	if !errors.As(err, &checkErr) {
		t.Fatal("expected CheckError type")
	}

	if checkErr.Category != ErrorCategoryResponseHeader {
		t.Errorf("Category = %v, want %v", checkErr.Category, ErrorCategoryResponseHeader)
	}
}

func TestNewDatabaseError(t *testing.T) {
	err := NewDatabaseError("expected %d rows, got %d", 5, 3)

	var checkErr *CheckError
	if !errors.As(err, &checkErr) {
		t.Fatal("expected CheckError type")
	}

	if checkErr.Category != ErrorCategoryDatabase {
		t.Errorf("Category = %v, want %v", checkErr.Category, ErrorCategoryDatabase)
	}
}

func TestNewDatabaseErrorWithCause(t *testing.T) {
	cause := errors.New("connection lost")
	err := NewDatabaseErrorWithCause(cause, "failed to query database")

	var checkErr *CheckError
	if !errors.As(err, &checkErr) {
		t.Fatal("expected CheckError type")
	}

	if checkErr.Category != ErrorCategoryDatabase {
		t.Errorf("Category = %v, want %v", checkErr.Category, ErrorCategoryDatabase)
	}

	if !errors.Is(err, cause) {
		t.Error("expected error to wrap cause")
	}
}

func TestNewMockError(t *testing.T) {
	err := NewMockError("at path %s: number of calls does not match", "/api/users")

	var checkErr *CheckError
	if !errors.As(err, &checkErr) {
		t.Fatal("expected CheckError type")
	}

	if checkErr.Category != ErrorCategoryMock {
		t.Errorf("Category = %v, want %v", checkErr.Category, ErrorCategoryMock)
	}
}

func TestErrorCategories(t *testing.T) {
	categories := []ErrorCategory{
		ErrorCategoryStatusCode,
		ErrorCategoryResponseBody,
		ErrorCategoryResponseHeader,
		ErrorCategoryDatabase,
		ErrorCategoryMock,
	}

	expectedValues := []string{
		"status_code",
		"body",
		"header",
		"database",
		"mock",
	}

	for i, cat := range categories {
		if string(cat) != expectedValues[i] {
			t.Errorf("Category %d: got %q, want %q", i, cat, expectedValues[i])
		}
	}
}
