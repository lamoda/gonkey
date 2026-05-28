package allure_report

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lamoda/gonkey/mocks"
	"github.com/lamoda/gonkey/models"
)

func TestGroupErrorsByEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		errs           []error
		wantKeys       []string
		wantCountByKey map[string]int
	}{
		{
			name: "happy path",
			errs: []error{
				&mocks.RequestConstraintError{Endpoint: "/api/users"},
				&mocks.RequestConstraintError{Endpoint: "/api/orders"},
				&mocks.RequestConstraintError{Endpoint: "/api/users"},
			},
			wantKeys:       []string{"/api/users", "/api/orders"},
			wantCountByKey: map[string]int{"/api/users": 2, "/api/orders": 1},
		},
		{
			name:           "return empty map for nil errors",
			errs:           nil,
			wantKeys:       []string{},
			wantCountByKey: map[string]int{},
		},
		{
			name: "group errors without endpoints under empty key",
			errs: []error{
				errors.New("calls mismatch"),
				&mocks.RequestConstraintError{Endpoint: ""},
			},
			wantKeys:       []string{""},
			wantCountByKey: map[string]int{"": 2},
		},
		{
			name: "separate errors with and without endpoints",
			errs: []error{
				&mocks.RequestConstraintError{Endpoint: "/api/pay"},
				errors.New("unexpected error"),
				&mocks.RequestConstraintError{Endpoint: "/api/pay"},
			},
			wantKeys:       []string{"/api/pay", ""},
			wantCountByKey: map[string]int{"/api/pay": 2, "": 1},
		},
		{
			name: "extract URI from JSON path format",
			errs: []error{
				&mocks.RequestConstraintError{Endpoint: "$.uriVary./orders.accept"},
				&mocks.RequestConstraintError{Endpoint: "$.uriVary./users.list"},
			},
			wantKeys:       []string{"/orders.accept", "/users.list"},
			wantCountByKey: map[string]int{"/orders.accept": 1, "/users.list": 1},
		},
		{
			name: "group CallsMismatchError by endpoint",
			errs: []error{
				&mocks.CallsMismatchError{Path: "$.uriVary./orders", Expected: 2, Actual: 1},
				&mocks.CallsMismatchError{Path: "$", Expected: 3, Actual: 1},
			},
			wantKeys:       []string{"/orders", ""},
			wantCountByKey: map[string]int{"/orders": 1, "": 1},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := groupErrorsByEndpoint(tt.errs)

			assert.Len(t, result, len(tt.wantKeys))
			assert.Equal(t, tt.wantCountByKey, countErrorsByKey(result))
		})
	}
}

func countErrorsByKey(m map[string][]error) map[string]int {
	counts := make(map[string]int, len(m))
	for k, v := range m {
		counts[k] = len(v)
	}
	return counts
}

func TestFormatConstraintError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		err          error
		wantContains string
	}{
		{
			name:         "happy path",
			err:          errors.New("calls mismatch: expected 2, actual 1"),
			wantContains: "calls mismatch: expected 2, actual 1",
		},
		{
			name:         "extract message from mock error",
			err:          models.NewMockErrorWithService("payment", "unexpected request"),
			wantContains: "unexpected request",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := formatConstraintError(tt.err)
			assert.Contains(t, result, tt.wantContains)
		})
	}
}

func TestCategorizeErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		errs       []error
		wantCounts map[models.ErrorCategory]int
	}{
		{
			name: "happy path",
			errs: []error{
				models.NewMockErrorWithService("service", "mock error"),
				models.NewBodyError("body error"),
				models.NewStatusCodeError(200, 404),
			},
			wantCounts: map[models.ErrorCategory]int{
				models.ErrorCategoryMock:         1,
				models.ErrorCategoryResponseBody: 1,
				models.ErrorCategoryStatusCode:   1,
			},
		},
		{
			name:       "return empty map for nil errors",
			errs:       nil,
			wantCounts: map[models.ErrorCategory]int{},
		},
		{
			name: "group mock errors by category",
			errs: []error{
				models.NewMockErrorWithService("service1", "error 1"),
				models.NewMockErrorWithService("service1", "error 2"),
				models.NewMockErrorWithService("service2", "error 3"),
			},
			wantCounts: map[models.ErrorCategory]int{
				models.ErrorCategoryMock: 3,
			},
		},
		{
			name: "categorize body errors",
			errs: []error{
				models.NewBodyError("field mismatch"),
			},
			wantCounts: map[models.ErrorCategory]int{
				models.ErrorCategoryResponseBody: 1,
			},
		},
		{
			name: "treat uncategorized errors as body errors",
			errs: []error{
				errors.New("unknown error"),
			},
			wantCounts: map[models.ErrorCategory]int{
				models.ErrorCategoryResponseBody: 1,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := categorizeErrors(tt.errs)

			assert.Equal(t, tt.wantCounts, countErrorsByCategory(result))
		})
	}
}

func countErrorsByCategory(m map[models.ErrorCategory]ErrorsByIdentifier) map[models.ErrorCategory]int {
	counts := make(map[models.ErrorCategory]int, len(m))
	for category, byIdentifier := range m {
		for _, errs := range byIdentifier {
			counts[category] += len(errs)
		}
	}
	return counts
}

func TestCategorizeErrors_GroupByServiceIdentifier(t *testing.T) {
	t.Parallel()

	errs := []error{
		models.NewMockErrorWithService("service1", "error 1"),
		models.NewMockErrorWithService("service1", "error 2"),
		models.NewMockErrorWithService("service2", "error 3"),
	}

	result := categorizeErrors(errs)
	mockErrors := result[models.ErrorCategoryMock]

	assert.Len(t, mockErrors["service1"], 2, "service1 should have 2 errors")
	assert.Len(t, mockErrors["service2"], 1, "service2 should have 1 error")
}

func TestExtractURIFromPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "uriVary with leading slash",
			path: "$.uriVary./rpc/orders.accept",
			want: "/rpc/orders.accept",
		},
		{
			name: "uriVary with trailing slash only",
			path: "$.uriVary.RU210924-188956/",
			want: "RU210924-188956/",
		},
		{
			name: "uriVary with both leading and trailing slash",
			path: "$.uriVary./api/quick-checkouts/",
			want: "/api/quick-checkouts/",
		},
		{
			name: "uriVary without any slash",
			path: "$.uriVary.payments.validate",
			want: "payments.validate",
		},
		{
			name: "methodVary",
			path: "$.methodVary.POST",
			want: "POST",
		},
		{
			name: "root path returns empty",
			path: "$",
			want: "",
		},
		{
			name: "fallback for unknown format with slash",
			path: "unknown./api/test",
			want: "/api/test",
		},
		{
			name: "fallback for unknown format without slash",
			path: "unknown.path",
			want: "",
		},
		{
			name: "sequence fallback index 0",
			path: "$.0",
			want: "step 0",
		},
		{
			name: "sequence fallback index 1",
			path: "$.1",
			want: "step 1",
		},
		{
			name: "sequence fallback index 10",
			path: "$.10",
			want: "step 10",
		},
		{
			name: "basedOnRequest prefix index 0",
			path: "$.basedOnRequest.0",
			want: "case 0",
		},
		{
			name: "basedOnRequest prefix index 2",
			path: "$.basedOnRequest.2",
			want: "case 2",
		},
		{
			name: "sequence prefix index 0",
			path: "$.sequence.0",
			want: "step 0",
		},
		{
			name: "sequence prefix index 1",
			path: "$.sequence.1",
			want: "step 1",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := extractURIFromPath(tt.path)
			assert.Equal(t, tt.want, result)
		})
	}
}
