package allure_report

import (
	"bytes"
	"encoding/xml"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/lamoda/gonkey/output/allure_report/beans"
)

type Allure struct {
	Suites    []*beans.Suite
	TargetDir string
}

func New(suites []*beans.Suite) *Allure {
	return &Allure{Suites: suites, TargetDir: "allure-results"}
}

func (a *Allure) GetCurrentSuite() *beans.Suite {
	return a.Suites[0]
}

func (a *Allure) StartSuite(name string, start time.Time) {
	a.Suites = append(a.Suites, beans.NewSuite(name, start))
}

func (a *Allure) EndSuite(end time.Time) error {
	suite := a.GetCurrentSuite()
	suite.SetEnd(end)
	if suite.HasTests() {
		if err := writeSuite(a.TargetDir, suite); err != nil {
			return err
		}
	}
	// remove first/current suite
	a.Suites = a.Suites[1:]

	return nil
}

var (
	currentState = map[*beans.Suite]*beans.TestCase{}
	currentStep  = map[*beans.Suite]*beans.Step{}
)

func (a *Allure) StartCase(testName string, start time.Time) *beans.TestCase {
	test := beans.NewTestCase(testName, start)
	step := beans.NewStep(testName, start)
	suite := a.GetCurrentSuite()
	currentState[suite] = test
	currentStep[suite] = step
	suite.AddTest(test)

	return test
}

func (a *Allure) EndCase(status string, err error, end time.Time) {
	suite := a.GetCurrentSuite()
	test, ok := currentState[suite]
	if ok {
		test.End(status, err, end)
	}
}

func (a *Allure) CreateStep(name string, stepFunc func()) {
	status := `passed`
	a.StartStep(name, time.Now())
	// if test error
	stepFunc()
	// end
	a.EndStep(status, time.Now())
}

func (a *Allure) StartStep(stepName string, start time.Time) {
	var (
		// FIXME: step is overwritten below
		step  = beans.NewStep(stepName, start)
		suite = a.GetCurrentSuite()
	)
	step = currentStep[suite]
	step.Parent.AddStep(step)
	currentStep[suite] = step
}

func (a *Allure) EndStep(status string, end time.Time) {
	suite := a.GetCurrentSuite()
	currentStep[suite].End(status, end)
	currentStep[suite] = currentStep[suite].Parent
}

func (a *Allure) AddAttachment(attachmentName, buf bytes.Buffer, typ string) {
	mime, ext := getBufferInfo(buf, typ)
	name, _ := writeBuffer(a.TargetDir, buf, ext)
	currentState[a.GetCurrentSuite()].AddAttachment(beans.NewAttachment(
		attachmentName.String(),
		mime,
		name,
		buf.Len()))
}

func (a *Allure) PendingCase(testName string, start time.Time) {
	a.StartCase(testName, start)
	a.EndCase("pending", errors.New("test ignored"), start)
}

// utils
func getBufferInfo(buf bytes.Buffer, typ string) (string, string) {
	//    exts,err := mime.ExtensionsByType(typ)
	//    if err != nil {
	//        mime.ParseMediaType()
	//    }
	return "text/plain", "txt"
}

func writeBuffer(pathDir string, buf bytes.Buffer, ext string) (string, error) {
	fileName := uuid.New().String() + `-attachment.` + ext
	err := os.WriteFile(filepath.Join(pathDir, fileName), buf.Bytes(), 0o644)

	return fileName, err
}

func writeSuite(pathDir string, suite *beans.Suite) error {
	b, err := xml.Marshal(suite)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(pathDir, uuid.New().String()+`-testsuite.xml`), b, 0o644)
	if err != nil {
		return err
	}

	return nil
}
