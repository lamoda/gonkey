package mocks

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lamoda/gonkey/models"
)

func TestError_Unwrap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		innerErr       error
		serviceName    string
		wantCategory   models.ErrorCategory
		wantIdentifier string
	}{
		{
			name:           "unwrap CheckError with service name",
			innerErr:       models.NewMockErrorWithService("test-service", "test error"),
			serviceName:    "test-service",
			wantCategory:   models.ErrorCategoryMock,
			wantIdentifier: "test-service",
		},
		{
			name:           "unwrap CheckError without service name",
			innerErr:       models.NewMockError("test error without service"),
			serviceName:    "wrapper-service",
			wantCategory:   models.ErrorCategoryMock,
			wantIdentifier: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			wrappedErr := &Error{
				error:       tt.innerErr,
				ServiceName: tt.serviceName,
			}

			var checkErr *models.CheckError
			require.True(t, errors.As(wrappedErr, &checkErr), "errors.As should extract CheckError from mocks.Error")

			assert.Equal(t, tt.wantCategory, checkErr.GetCategory())
			assert.Equal(t, tt.wantIdentifier, checkErr.GetIdentifier())
		})
	}
}

func TestRequestConstraintError_Unwrap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		innerErr       error
		wantCategory   models.ErrorCategory
		wantIdentifier string
	}{
		{
			name:           "unwrap CheckError with service name",
			innerErr:       models.NewMockErrorWithService("test-service", "constraint failed"),
			wantCategory:   models.ErrorCategoryMock,
			wantIdentifier: "test-service",
		},
		{
			name:           "unwrap CheckError without service name",
			innerErr:       models.NewMockError("constraint failed"),
			wantCategory:   models.ErrorCategoryMock,
			wantIdentifier: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			wrappedErr := &RequestConstraintError{
				error: tt.innerErr,
			}

			var checkErr *models.CheckError
			require.True(t, errors.As(wrappedErr, &checkErr), "errors.As should extract CheckError from RequestConstraintError")

			assert.Equal(t, tt.wantCategory, checkErr.GetCategory())
			assert.Equal(t, tt.wantIdentifier, checkErr.GetIdentifier())
		})
	}
}

func TestError_UnwrapChain(t *testing.T) {
	t.Parallel()

	checkErr := models.NewMockErrorWithService("my-service", "deep error")

	mocksErr := &Error{
		error:       checkErr,
		ServiceName: "my-service",
	}

	constraintErr := &RequestConstraintError{
		error: mocksErr,
	}

	var extractedErr *models.CheckError
	require.True(t, errors.As(constraintErr, &extractedErr), "errors.As should extract CheckError from deeply wrapped error")

	assert.Equal(t, models.ErrorCategoryMock, extractedErr.GetCategory())
	assert.Equal(t, "my-service", extractedErr.GetIdentifier())
}
