package allure_report

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatMockBrief(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		info mockInfo
		want string
	}{
		{
			name: "happy path",
			info: mockInfo{
				ServiceName:   "PaymentService",
				Strategy:      "constant",
				EndpointCount: 1,
			},
			want: "PaymentService (constant)",
		},
		{
			name: "return strategy with endpoints count for uriVary",
			info: mockInfo{
				ServiceName:   "OrderService",
				Strategy:      "uriVary",
				EndpointCount: 3,
			},
			want: "OrderService (uriVary, 3 endpoints)",
		},
		{
			name: "return strategy without count when uriVary has no endpoints",
			info: mockInfo{
				ServiceName:   "OrderService",
				Strategy:      "uriVary",
				EndpointCount: 0,
			},
			want: "OrderService (uriVary)",
		},
		{
			name: "return strategy with methods count for methodVary",
			info: mockInfo{
				ServiceName:   "ApiGateway",
				Strategy:      "methodVary",
				EndpointCount: 2,
			},
			want: "ApiGateway (methodVary, 2 methods)",
		},
		{
			name: "return strategy with steps count for sequence",
			info: mockInfo{
				ServiceName: "NotificationService",
				Strategy:    "sequence",
				StepsCount:  5,
			},
			want: "NotificationService (sequence, 5 steps)",
		},
		{
			name: "return strategy without count when sequence has no steps",
			info: mockInfo{
				ServiceName: "NotificationService",
				Strategy:    "sequence",
				StepsCount:  0,
			},
			want: "NotificationService (sequence)",
		},
		{
			name: "return strategy with cases count for basedOnRequest",
			info: mockInfo{
				ServiceName:   "RouterService",
				Strategy:      "basedOnRequest",
				EndpointCount: 4,
			},
			want: "RouterService (basedOnRequest, 4 cases)",
		},
		{
			name: "return only strategy for nop",
			info: mockInfo{
				ServiceName: "IgnoredService",
				Strategy:    "nop",
			},
			want: "IgnoredService (nop)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := formatMockBrief(tt.info)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractMocksInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		mocks map[string]interface{}
		want  []mockInfo
	}{
		{
			name:  "return nil for empty mocks",
			mocks: nil,
			want:  nil,
		},
		{
			name: "happy path",
			mocks: map[string]interface{}{
				"service1": map[interface{}]interface{}{
					"strategy": "constant",
					"path":     "/api/test",
				},
			},
			want: []mockInfo{
				{
					ServiceName:   "service1",
					Strategy:      "constant",
					Endpoints:     []string{"/api/test"},
					EndpointCount: 1,
				},
			},
		},
		{
			name: "extract steps count for sequence strategy",
			mocks: map[string]interface{}{
				"seqService": map[interface{}]interface{}{
					"strategy": "sequence",
					"sequence": []interface{}{
						map[interface{}]interface{}{"body": "{}"},
						map[interface{}]interface{}{"body": "{}"},
						map[interface{}]interface{}{"body": "{}"},
					},
				},
			},
			want: []mockInfo{
				{
					ServiceName:   "seqService",
					Strategy:      "sequence",
					Endpoints:     []string{"step 0", "step 1", "step 2"},
					EndpointCount: 3,
					StepsCount:    3,
				},
			},
		},
		{
			name: "extract endpoints for uriVary strategy",
			mocks: map[string]interface{}{
				"multiService": map[interface{}]interface{}{
					"strategy": "uriVary",
					"uris": map[interface{}]interface{}{
						"/api/users":  map[interface{}]interface{}{},
						"/api/orders": map[interface{}]interface{}{},
					},
				},
			},
			want: []mockInfo{
				{
					ServiceName:   "multiService",
					Strategy:      "uriVary",
					Endpoints:     []string{"/api/orders", "/api/users"},
					EndpointCount: 2,
				},
			},
		},
		{
			name: "sort services alphabetically",
			mocks: map[string]interface{}{
				"zService": map[interface{}]interface{}{
					"strategy": "constant",
				},
				"aService": map[interface{}]interface{}{
					"strategy": "template",
				},
			},
			want: []mockInfo{
				{
					ServiceName:   "aService",
					Strategy:      "template",
					Endpoints:     []string{"/"},
					EndpointCount: 1,
				},
				{
					ServiceName:   "zService",
					Strategy:      "constant",
					Endpoints:     []string{"/"},
					EndpointCount: 1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := extractMocksInfo(tt.mocks)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractEndpoints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		strategy string
		def      map[interface{}]interface{}
		want     []string
	}{
		{
			name:     "happy path",
			strategy: "constant",
			def:      map[interface{}]interface{}{"path": "/api/users"},
			want:     []string{"/api/users"},
		},
		{
			name:     "return root path when constant has no path",
			strategy: "constant",
			def:      map[interface{}]interface{}{},
			want:     []string{"/"},
		},
		{
			name:     "return filename for file strategy",
			strategy: "file",
			def:      map[interface{}]interface{}{"filename": "/path/to/response.json"},
			want:     []string{"file: response.json"},
		},
		{
			name:     "return dash for nop strategy",
			strategy: "nop",
			def:      map[interface{}]interface{}{},
			want:     []string{"-"},
		},
		{
			name:     "return dash for dropRequest strategy",
			strategy: "dropRequest",
			def:      map[interface{}]interface{}{},
			want:     []string{"-"},
		},
		{
			name:     "return methods for methodVary strategy",
			strategy: "methodVary",
			def: map[interface{}]interface{}{
				"methods": map[interface{}]interface{}{
					"GET":  map[interface{}]interface{}{},
					"POST": map[interface{}]interface{}{},
				},
			},
			want: []string{"GET", "POST"},
		},
		{
			name:     "return uris without basePath for uriVary strategy",
			strategy: "uriVary",
			def: map[interface{}]interface{}{
				"basePath": "/api/v1",
				"uris": map[interface{}]interface{}{
					"/users":  map[interface{}]interface{}{},
					"/orders": map[interface{}]interface{}{},
				},
			},
			want: []string{"/orders", "/users"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := extractEndpoints(tt.strategy, tt.def)
			assert.Equal(t, tt.want, got)
		})
	}
}
