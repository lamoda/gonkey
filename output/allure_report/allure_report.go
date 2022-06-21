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

	for i, dbresult := range result.DatabaseResult {
		if dbresult.Query != "" {
			o.allure.AddAttachment(
				*bytes.NewBufferString(fmt.Sprintf("Db Query #%d", i+1)),
				*bytes.NewBufferString(fmt.Sprintf(`SQL string: %s`, dbresult.Query)),
				"txt")
			o.allure.AddAttachment(
				*bytes.NewBufferString(fmt.Sprintf("Db Response #%d", i+1)),
				*bytes.NewBufferString(fmt.Sprintf(`Respone: %s`, dbresult.Response)),
				"txt")
		}
	}

	status, err := result.AllureStatus()
	o.allure.EndCase(status, err, time.Now())

	return nil
}

func (o *AllureReportOutput) Finalize() {
	o.allure.EndSuite(time.Now())
}
