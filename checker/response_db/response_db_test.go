package response_db

import "testing"

func TestCompareDbRespWithoutOrdering(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
		actual   []string
		fail     bool
	}{
		{
			name:     "one line",
			expected: []string{"{ \"name\": \"John\" }"},
			actual:   []string{"{ \"name\": \"John\" }"},
			fail:     false,
		},
		{
			name:     "two lines",
			expected: []string{"{ \"name\": \"John\" }", "{ \"surname\": \"Doe\" }"},
			actual:   []string{"{ \"name\": \"John\" }", "{ \"surname\": \"Doe\" }"},
			fail:     false,
		},
		{
			name:     "two lines; different order",
			expected: []string{"{ \"surname\": \"Doe\" }", "{ \"name\": \"John\" }"},
			actual:   []string{"{ \"name\": \"John\" }", "{ \"surname\": \"Doe\" }"},
			fail:     false,
		},
		{
			name:     "error",
			expected: []string{"{ \"name\": \"Jane\" }", "{ \"surname\": \"Doe\" }"},
			actual:   []string{"{ \"name\": \"John\" }", "{ \"surname\": \"Doe\" }"},
			fail:     true,
		},
	}

	for _, tt := range tests {
		errors, err := compareDbRespWithoutOrdering(tt.expected, tt.actual, tt.name)
		if tt.fail {
			if err == nil && len(errors) == 0 {
				t.Errorf("expected errors")
			}
		} else {
			if err != nil {
				t.Error(err)
			}
			if len(errors) > 0 {
				t.Errorf("got errors")
			}
		}
	}
}

func TestCompareDbResp(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
		actual   []string
		fail     bool
	}{
		{
			name:     "one line",
			expected: []string{"{ \"name\": \"John\" }"},
			actual:   []string{"{ \"name\": \"John\" }"},
			fail:     false,
		},
		{
			name:     "two lines",
			expected: []string{"{ \"name\": \"John\" }", "{ \"surname\": \"Doe\" }"},
			actual:   []string{"{ \"name\": \"John\" }", "{ \"surname\": \"Doe\" }"},
			fail:     false,
		},
		{
			name:     "two lines; different order",
			expected: []string{"{ \"surname\": \"Doe\" }", "{ \"name\": \"John\" }"},
			actual:   []string{"{ \"name\": \"John\" }", "{ \"surname\": \"Doe\" }"},
			fail:     true,
		},
		{
			name:     "error",
			expected: []string{"{ \"name\": \"Jane\" }", "{ \"surname\": \"Doe\" }"},
			actual:   []string{"{ \"name\": \"John\" }", "{ \"surname\": \"Doe\" }"},
			fail:     true,
		},
	}

	for _, tt := range tests {
		errors, err := compareDbResp(tt.expected, tt.actual, tt.name, "")
		if tt.fail {
			if err == nil && len(errors) == 0 {
				t.Errorf("expected errors")
			}
		} else {
			if err != nil {
				t.Error(err)
			}
			if len(errors) > 0 {
				t.Errorf("got errors")
			}
		}
	}
}
