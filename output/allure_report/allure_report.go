package allure_report

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lamoda/gonkey/models"
	"github.com/lamoda/gonkey/output"
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
	if result.DbQuery != "" {
		o.allure.AddAttachment(
			*bytes.NewBufferString("Db Query"),
			*bytes.NewBufferString(fmt.Sprintf(`SQL string: %s`, result.DbQuery)),
			"txt")
		o.allure.AddAttachment(
			*bytes.NewBufferString("Db Response"),
			*bytes.NewBufferString(fmt.Sprintf(`Response: %s`, result.DbResponse)),
			"txt")
	}

	status, err := result.AllureStatus()
	o.allure.EndCase(status, err, time.Now())

	return nil
}

func (o *AllureReportOutput) Finalize() {
	_ = o.allure.EndSuite(time.Now())
}
