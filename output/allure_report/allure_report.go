package allure_report

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/keyclaim/gonkey/models"
	"github.com/keyclaim/gonkey/output"
)

type AllureReportOutput struct {
	output.OutputInterface

	reportLocation string
	allure         Allure
}

func NewOutput(suiteName, reportLocation string) *AllureReportOutput {
	resultsDir, _ := filepath.Abs(reportLocation)
	if err := os.Mkdir(resultsDir, 0777); err != nil {
		// likely dir is already exists
	}
	a := Allure{
		Suites:    nil,
		TargetDir: resultsDir,
	}
	a.StartSuite(suiteName, time.Now())
	return &AllureReportOutput{
		reportLocation: reportLocation,
		allure:         a,
	}
}

func (o *AllureReportOutput) Process(t models.TestInterface, result *models.Result) error {
	testCase := o.allure.StartCase(t.GetName(), time.Now())
	testCase.AddLabel("story", result.Path)
	o.allure.AddAttachment(
		*bytes.NewBufferString("Request"),
		*bytes.NewBufferString(fmt.Sprintf(`Query: %s \n Body: %s`, result.Query, result.RequestBody)),
		"txt")
	o.allure.AddAttachment(
		*bytes.NewBufferString("Response"),
		*bytes.NewBufferString(fmt.Sprintf(`Body: %s`, result.ResponseBody)),
		"txt")
	if !result.Passed() {
		ers := ""
		for _, e := range result.Errors {
			ers = ers + e.Error() + "\n"
		}
		o.allure.EndCase("failed", errors.New(ers), time.Now())
	} else {
		o.allure.EndCase("passed", nil, time.Now())
	}
	return nil
}

func (o *AllureReportOutput) Finalize() {
	o.allure.EndSuite(time.Now())
}
