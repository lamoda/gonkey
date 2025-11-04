package response_header

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lamoda/gonkey/models"
	"github.com/lamoda/gonkey/testloader/yaml_file"
)

func TestCheckShouldMatchSubset(t *testing.T) {
	test := &yaml_file.Test{
		ResponseHeaders: map[int]map[string]string{
			200: {
				"content-type": "application/json",
				"ACCEPT":       "text/html",
			},
		},
	}

	result := &models.Result{
		ResponseStatusCode: 200,
		ResponseHeaders: map[string][]string{
			"Content-Type": {
				"application/json",
			},
			"Accept": {
				// uts enough for expected value to match only one entry of the actual values slice
				"application/json",
				"text/html",
			},
		},
	}

	checker := NewChecker()
	errs, err := checker.Check(test, result)

	assert.NoError(t, err, "Check must not result with an error")
	assert.Empty(t, errs, "Check must succeed")
}

func TestCheckWhenNotMatchedShouldReturnError(t *testing.T) {
	test := &yaml_file.Test{
		ResponseHeaders: map[int]map[string]string{
			200: {
				"content-type": "application/json",
				"accept":       "text/html",
			},
		},
	}

	result := &models.Result{
		ResponseStatusCode: 200,
		ResponseHeaders: map[string][]string{
			// no header "Content-Type" in response
			"Accept": {
				"application/json",
			},
		},
	}

	checker := NewChecker()
	errs, err := checker.Check(test, result)

	sort.Slice(errs, func(i, j int) bool {
		return errs[i].Error() < errs[j].Error()
	})

	assert.NoError(t, err, "Check must not result with an error")
	assert.Len(t, errs, 2)

	// Verify errors are typed as CheckError
	assert.Contains(t, errs[0].Error(), "response does not include expected header Content-Type")
	assert.Contains(t, errs[1].Error(), "response header Accept value does not match expected text/html")
}
