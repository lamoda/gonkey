package allure_report

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lamoda/gonkey/models"
)

type AllureReportOutput struct {
	reportLocation string
	allure         Allure
}

func NewOutput(suiteName, reportLocation string) *AllureReportOutput {
	resultsDir, _ := filepath.Abs(reportLocation)
	_ = os.Mkdir(resultsDir, 0o777)
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
	testCase.SetDescriptionOrDefaultValue(t.GetDescription(), "No description")
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
				*bytes.NewBufferString(fmt.Sprintf(`Response: %s`, dbresult.Response)),
				"txt")
		}
	}

	status, err := result.AllureStatus()
	o.allure.EndCase(status, err, time.Now())

	return nil
}

func (o *AllureReportOutput) Finalize() {
	_ = o.allure.EndSuite(time.Now())
}
