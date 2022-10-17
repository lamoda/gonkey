package runner

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lamoda/gonkey/models"
)

func TestConsoleHandler_HandleTest_Summary(t *testing.T) {
	type result struct {
		result *models.Result
		err    error
	}
	tests := []struct {
		name            string
		testExecResults []result
		wantErr         bool
		expectedSummary *models.Summary
	}{
		{
			name:            "1 successful test",
			testExecResults: []result{{result: &models.Result{Errors: nil}}},
			expectedSummary: &models.Summary{
				Success: true,
				Failed:  0,
				Total:   1,
			},
		},
		{
			name: "1 successful test, 1 broken test",
			testExecResults: []result{
				{result: &models.Result{}, err: nil},
				{result: &models.Result{}, err: errTestBroken},
			},
			expectedSummary: &models.Summary{
				Success: true,
				Failed:  0,
				Broken:  1,
				Total:   2,
			},
		},
		{
			name: "1 successful test, 1 skipped test",
			testExecResults: []result{
				{result: &models.Result{}, err: nil},
				{result: &models.Result{}, err: errTestSkipped},
			},
			expectedSummary: &models.Summary{
				Success: true,
				Skipped: 1,
				Total:   2,
			},
		},
		{
			name: "1 successful test, 1 failed test",
			testExecResults: []result{
				{result: &models.Result{}, err: nil},
				{result: &models.Result{Errors: []error{errors.New("some err")}}, err: nil},
			},
			expectedSummary: &models.Summary{
				Success: false,
				Failed:  1,
				Total:   2,
			},
		},
		{
			name: "test with unexpected error",
			testExecResults: []result{
				{result: &models.Result{}, err: errors.New("unexpected error")},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewConsoleHandler()

			for _, execResult := range tt.testExecResults {
				executor := func(models.TestInterface) (*models.Result, error) {
					return execResult.result, execResult.err
				}
				err := h.HandleTest(nil, executor)
				if tt.wantErr {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
			}

			require.Equal(t, tt.expectedSummary, h.Summary())
		})
	}
}
