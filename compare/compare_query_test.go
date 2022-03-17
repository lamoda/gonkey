package compare

import "testing"

func TestCompareQuery(t *testing.T) {
	tests := []struct {
		name          string
		expectedQuery []string
		actualQuery   []string
	}{
		{
			name:          "simple expected and actual",
			expectedQuery: []string{"cake"},
			actualQuery:   []string{"cake"},
		},
		{
			name:          "expected and actual with two values",
			expectedQuery: []string{"cake", "tea"},
			actualQuery:   []string{"cake", "tea"},
		},
		{
			name:          "expected and actual with two values and different order",
			expectedQuery: []string{"cake", "tea"},
			actualQuery:   []string{"tea", "cake"},
		},
		{
			name:          "expected and actual with same values",
			expectedQuery: []string{"tea", "cake", "tea"},
			actualQuery:   []string{"cake", "tea", "tea"},
		},
		{
			name:          "expected and actual with regexp",
			expectedQuery: []string{"tea", "$matchRegexp(^c\\w+)"},
			actualQuery:   []string{"cake", "tea"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := CompareQuery(tt.expectedQuery, tt.actualQuery)
			if err != nil {
				t.Error(err)
			}
			if !ok {
				t.Errorf("expected and actual queries do not match")
			}
		})
	}
}
