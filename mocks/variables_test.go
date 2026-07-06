package mocks

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lamoda/gonkey/compare"
)

func execDefinition(t *testing.T, def *Definition, method, target, body string) []error {
	t.Helper()

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	r := httptest.NewRequest(method, target, bodyReader)
	w := httptest.NewRecorder()

	return def.Execute(w, r)
}

func Test_Definition_CaptureVariables(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		template string
		method   string
		target   string
		body     string
		want     map[string]string
	}{
		{
			name:     "captures field from json body",
			varName:  "userId",
			template: `{{ .request.Json.user.id }}`,
			method:   "POST",
			target:   "/",
			body:     `{"user": {"id": "42"}}`,
			want:     map[string]string{"userId": "42"},
		},
		{
			name:     "captures query parameter",
			varName:  "page",
			template: `{{ .request.Query "page" }}`,
			method:   "GET",
			target:   "/list?page=7",
			body:     "",
			want:     map[string]string{"page": "7"},
		},
		{
			name:     "captures request header",
			varName:  "requestId",
			template: `{{ .request.Header "X-Request-Id" }}`,
			method:   "GET",
			target:   "/",
			body:     "",
			want:     map[string]string{"requestId": "req-1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext, err := newVariableExtractor(tt.varName, tt.template, "")
			require.NoError(t, err)

			def := NewDefinition("$", nil, NewConstantReplyWithCode([]byte("ok"), 200, nil), CallsNoConstraint)
			def.variablesToSet = []*variableExtractor{ext}

			var bodyReader io.Reader
			if tt.body != "" {
				bodyReader = strings.NewReader(tt.body)
			}
			r := httptest.NewRequest(tt.method, tt.target, bodyReader)
			r.Header.Set("X-Request-Id", "req-1")
			w := httptest.NewRecorder()

			errs := def.Execute(w, r)

			require.Empty(t, errs)
			assert.Equal(t, tt.want, def.CapturedVariables())
		})
	}
}

func Test_Definition_CaptureVariables_FromCall(t *testing.T) {
	tests := []struct {
		name     string
		fromCall string
		want     string
	}{
		{
			name:     "last call wins by default",
			fromCall: "",
			want:     "second",
		},
		{
			name:     "last call wins explicitly",
			fromCall: "last",
			want:     "second",
		},
		{
			name:     "first call wins",
			fromCall: "first",
			want:     "first",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext, err := newVariableExtractor("val", `{{ .request.Json.val }}`, tt.fromCall)
			require.NoError(t, err)

			def := NewDefinition("$", nil, NewConstantReplyWithCode([]byte("ok"), 200, nil), CallsNoConstraint)
			def.variablesToSet = []*variableExtractor{ext}

			errs := execDefinition(t, def, "POST", "/", `{"val": "first"}`)
			require.Empty(t, errs)
			errs = execDefinition(t, def, "POST", "/", `{"val": "second"}`)
			require.Empty(t, errs)

			assert.Equal(t, map[string]string{"val": tt.want}, def.CapturedVariables())
		})
	}
}

func Test_newVariableExtractor_Errors(t *testing.T) {
	tests := []struct {
		name     string
		template string
		fromCall string
	}{
		{
			name:     "empty template",
			template: "",
			fromCall: "",
		},
		{
			name:     "malformed template",
			template: `{{ .request.Json.val `,
			fromCall: "",
		},
		{
			name:     "unknown from_call",
			template: `{{ .request.Json.val }}`,
			fromCall: "second",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newVariableExtractor("val", tt.template, tt.fromCall)
			assert.Error(t, err)
		})
	}
}

func Test_Definition_CaptureVariables_MissingJSONFieldFails(t *testing.T) {
	ext, err := newVariableExtractor("val", `{{ .request.Json.missing }}`, "")
	require.NoError(t, err)

	def := NewDefinition("$", nil, NewConstantReplyWithCode([]byte("ok"), 200, nil), CallsNoConstraint)
	def.variablesToSet = []*variableExtractor{ext}

	errs := execDefinition(t, def, "POST", "/", `{"val": "first"}`)

	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "val")
	assert.Empty(t, def.CapturedVariables())
}

func Test_Definition_CaptureVariables_SkippedWhenConstraintsFail(t *testing.T) {
	constraint, err := newBodyMatchesJSONConstraint(`{"expected": "body"}`, defaultCompareParams())
	require.NoError(t, err)

	ext, err := newVariableExtractor("val", `{{ .request.Json.val }}`, "")
	require.NoError(t, err)

	def := NewDefinition("$", []verifier{constraint}, NewConstantReplyWithCode([]byte("ok"), 200, nil), CallsNoConstraint)
	def.variablesToSet = []*variableExtractor{ext}

	errs := execDefinition(t, def, "POST", "/", `{"val": "first"}`)

	require.NotEmpty(t, errs)
	assert.Empty(t, def.CapturedVariables())
}

func Test_Definition_CaptureVariables_BodyStaysReadableForConstraintsAndTemplate(t *testing.T) {
	constraint, err := newBodyMatchesJSONConstraint(`{"val": "first"}`, defaultCompareParams())
	require.NoError(t, err)

	replyStrategy, err := newTemplateReply(`{"echo": "{{ .request.Json.val }}"}`, 200, nil)
	require.NoError(t, err)

	ext, err := newVariableExtractor("val", `{{ .request.Json.val }}`, "")
	require.NoError(t, err)

	def := NewDefinition("$", []verifier{constraint}, replyStrategy, CallsNoConstraint)
	def.variablesToSet = []*variableExtractor{ext}

	r := httptest.NewRequest("POST", "/", strings.NewReader(`{"val": "first"}`))
	w := httptest.NewRecorder()

	errs := def.Execute(w, r)

	require.Empty(t, errs)
	assert.Equal(t, `{"echo": "first"}`, w.Body.String())
	assert.Equal(t, map[string]string{"val": "first"}, def.CapturedVariables())
}

func Test_Definition_ResetRunningContext_ClearsCapturedVariables(t *testing.T) {
	ext, err := newVariableExtractor("val", `{{ .request.Json.val }}`, "")
	require.NoError(t, err)

	def := NewDefinition("$", nil, NewConstantReplyWithCode([]byte("ok"), 200, nil), CallsNoConstraint)
	def.variablesToSet = []*variableExtractor{ext}

	errs := execDefinition(t, def, "POST", "/", `{"val": "first"}`)
	require.Empty(t, errs)

	def.ResetRunningContext()

	assert.Empty(t, def.CapturedVariables())
}

func defaultCompareParams() compare.Params {
	return compare.Params{IgnoreArraysOrdering: true}
}

func execServiceMock(t *testing.T, svc *ServiceMock, body string) []error {
	t.Helper()

	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	svc.ServeHTTP(w, r)

	return svc.EndRunningContext()
}

func loadTestDefinition(t *testing.T, rawDef map[string]interface{}) *Definition {
	t.Helper()

	def, err := NewLoader(NewNop("service")).loadDefinition("$", rawDef)
	require.NoError(t, err)

	return def
}

func Test_Definition_CapturedVariables_UriVary(t *testing.T) {
	def := loadTestDefinition(t, map[string]interface{}{
		"strategy": "uriVary",
		"uris": map[string]interface{}{
			"/users": map[string]interface{}{
				"strategy": "constant",
				"body":     "ok",
				"variablesToSet": map[string]interface{}{
					"userId": `{{ .request.Json.id }}`,
				},
			},
			"/orders": map[string]interface{}{
				"strategy": "constant",
				"body":     "ok",
				"variablesToSet": map[string]interface{}{
					"orderId": `{{ .request.Json.id }}`,
				},
			},
		},
	})

	errs := execDefinition(t, def, "POST", "/users", `{"id": "u1"}`)
	require.Empty(t, errs)
	errs = execDefinition(t, def, "POST", "/orders", `{"id": "o1"}`)
	require.Empty(t, errs)

	assert.Equal(t, map[string]string{"userId": "u1", "orderId": "o1"}, def.CapturedVariables())
}

func Test_Definition_CapturedVariables_MethodVary(t *testing.T) {
	def := loadTestDefinition(t, map[string]interface{}{
		"strategy": "methodVary",
		"methods": map[string]interface{}{
			"POST": map[string]interface{}{
				"strategy": "constant",
				"body":     "ok",
				"variablesToSet": map[string]interface{}{
					"created": `{{ .request.Json.id }}`,
				},
			},
		},
	})

	errs := execDefinition(t, def, "POST", "/", `{"id": "c1"}`)
	require.Empty(t, errs)

	assert.Equal(t, map[string]string{"created": "c1"}, def.CapturedVariables())
}

func Test_Definition_CapturedVariables_Sequence(t *testing.T) {
	def := loadTestDefinition(t, map[string]interface{}{
		"strategy": "sequence",
		"sequence": []interface{}{
			map[string]interface{}{
				"strategy": "constant",
				"body":     "ok",
				"variablesToSet": map[string]interface{}{
					"firstCall": `{{ .request.Json.id }}`,
				},
			},
			map[string]interface{}{
				"strategy": "constant",
				"body":     "ok",
				"variablesToSet": map[string]interface{}{
					"secondCall": `{{ .request.Json.id }}`,
				},
			},
		},
	})

	errs := execDefinition(t, def, "POST", "/", `{"id": "1"}`)
	require.Empty(t, errs)
	errs = execDefinition(t, def, "POST", "/", `{"id": "2"}`)
	require.Empty(t, errs)

	assert.Equal(t, map[string]string{"firstCall": "1", "secondCall": "2"}, def.CapturedVariables())
}

func Test_Definition_CapturedVariables_BasedOnRequest(t *testing.T) {
	def := loadTestDefinition(t, map[string]interface{}{
		"strategy": "basedOnRequest",
		"uris": []interface{}{
			map[string]interface{}{
				"strategy": "constant",
				"body":     "ok",
				"requestConstraints": []interface{}{
					map[string]interface{}{
						"kind":          "queryMatches",
						"expectedQuery": "kind=user",
					},
				},
				"variablesToSet": map[string]interface{}{
					"userId": `{{ .request.Json.id }}`,
				},
			},
			map[string]interface{}{
				"strategy": "constant",
				"body":     "ok",
				"requestConstraints": []interface{}{
					map[string]interface{}{
						"kind":          "queryMatches",
						"expectedQuery": "kind=order",
					},
				},
				"variablesToSet": map[string]interface{}{
					"orderId": `{{ .request.Json.id }}`,
				},
			},
		},
	})

	errs := execDefinition(t, def, "POST", "/?kind=user", `{"id": "u1"}`)
	require.Empty(t, errs)
	errs = execDefinition(t, def, "POST", "/?kind=order", `{"id": "o1"}`)
	require.Empty(t, errs)

	assert.Equal(t, map[string]string{"userId": "u1", "orderId": "o1"}, def.CapturedVariables())
}

func Test_Mocks_CapturedVariables_AggregatesServices(t *testing.T) {
	m := NewNop("users", "orders")
	loader := NewLoader(m)
	err := loader.Load(map[string]interface{}{
		"users": map[string]interface{}{
			"strategy": "constant",
			"body":     "ok",
			"variablesToSet": map[string]interface{}{
				"userId": `{{ .request.Json.id }}`,
			},
		},
		"orders": map[string]interface{}{
			"strategy": "constant",
			"body":     "ok",
			"variablesToSet": map[string]interface{}{
				"orderId": `{{ .request.Json.id }}`,
			},
		},
	})
	require.NoError(t, err)

	errs := execServiceMock(t, m.Service("users"), `{"id": "u1"}`)
	require.Empty(t, errs)
	errs = execServiceMock(t, m.Service("orders"), `{"id": "o1"}`)
	require.Empty(t, errs)

	assert.Equal(t, map[string]string{"userId": "u1", "orderId": "o1"}, m.CapturedVariables())
}
