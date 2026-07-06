package mocks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Loader_LoadDefinition_VariablesToSet(t *testing.T) {
	tests := []struct {
		name           string
		rawDef         map[string]interface{}
		wantExtractors map[string]string // name -> fromCall
	}{
		{
			name: "short form",
			rawDef: map[string]interface{}{
				"strategy": "constant",
				"body":     "ok",
				"variablesToSet": map[string]interface{}{
					"userId": `{{ .request.Json.user.id }}`,
				},
			},
			wantExtractors: map[string]string{"userId": fromCallLast},
		},
		{
			name: "full form with fromCall",
			rawDef: map[string]interface{}{
				"strategy": "constant",
				"body":     "ok",
				"variablesToSet": map[string]interface{}{
					"userId": map[string]interface{}{
						"template": `{{ .request.Json.user.id }}`,
						"fromCall": "first",
					},
				},
			},
			wantExtractors: map[string]string{"userId": fromCallFirst},
		},
		{
			name: "full form defaults to last",
			rawDef: map[string]interface{}{
				"strategy": "constant",
				"body":     "ok",
				"variablesToSet": map[string]interface{}{
					"userId": map[string]interface{}{
						"template": `{{ .request.Json.user.id }}`,
					},
				},
			},
			wantExtractors: map[string]string{"userId": fromCallLast},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewLoader(NewNop("service"))

			def, err := loader.loadDefinition("$", tt.rawDef)

			require.NoError(t, err)
			got := map[string]string{}
			for _, e := range def.variablesToSet {
				got[e.name] = e.fromCall
			}
			assert.Equal(t, tt.wantExtractors, got)
		})
	}
}

func Test_Loader_LoadDefinition_VariablesToSet_Errors(t *testing.T) {
	tests := []struct {
		name            string
		variablesToSet  interface{}
		wantErrContains string
	}{
		{
			name:            "not a map",
			variablesToSet:  []interface{}{"userId"},
			wantErrContains: "`variablesToSet` requires map",
		},
		{
			name: "unknown fromCall",
			variablesToSet: map[string]interface{}{
				"userId": map[string]interface{}{
					"template": `{{ .request.Json.user.id }}`,
					"fromCall": "second",
				},
			},
			wantErrContains: "fromCall",
		},
		{
			name: "full form without template",
			variablesToSet: map[string]interface{}{
				"userId": map[string]interface{}{
					"fromCall": "first",
				},
			},
			wantErrContains: "template",
		},
		{
			name: "full form with unexpected key",
			variablesToSet: map[string]interface{}{
				"userId": map[string]interface{}{
					"template": `{{ .request.Json.user.id }}`,
					"calls":    1,
				},
			},
			wantErrContains: "unexpected key",
		},
		{
			name: "malformed template",
			variablesToSet: map[string]interface{}{
				"userId": `{{ .request.Json.user.id `,
			},
			wantErrContains: "template",
		},
		{
			name: "value is neither string nor map",
			variablesToSet: map[string]interface{}{
				"userId": 42,
			},
			wantErrContains: "userId",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewLoader(NewNop("service"))
			rawDef := map[string]interface{}{
				"strategy":       "constant",
				"body":           "ok",
				"variablesToSet": tt.variablesToSet,
			}

			_, err := loader.loadDefinition("$", rawDef)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrContains)
		})
	}
}
