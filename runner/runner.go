package runner

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/lamoda/gonkey/checker"
	"github.com/lamoda/gonkey/cmd_runner"
	"github.com/lamoda/gonkey/fixtures"
	"github.com/lamoda/gonkey/mocks"
	"github.com/lamoda/gonkey/models"
	"github.com/lamoda/gonkey/output"
	"github.com/lamoda/gonkey/testloader"
	"github.com/lamoda/gonkey/variables"
)

type Config struct {
	Host           string
	FixturesLoader fixtures.Loader
	Mocks          *mocks.Mocks
	MocksLoader    *mocks.Loader
	Variables      *variables.Variables
}

type Runner struct {
	loader   testloader.LoaderInterface
	output   []output.OutputInterface
	checkers []checker.CheckerInterface

	config *Config
}

func New(config *Config, loader testloader.LoaderInterface) *Runner {
	return &Runner{
		config: config,
		loader: loader,
	}
}

func (r *Runner) AddOutput(o ...output.OutputInterface) {
	r.output = append(r.output, o...)
}

func (r *Runner) AddCheckers(c ...checker.CheckerInterface) {
	r.checkers = append(r.checkers, c...)
}

func (r *Runner) Run() (*models.Summary, error) {
	if r.loader == nil {
		s := &models.Summary{
			Success: true,
			Failed:  0,
			Total:   0,
		}
		return s, nil
	}

	loader, err := r.loader.Load()
	if err != nil {
		return nil, err
	}

	tests := []models.TestInterface{}
	hasFocused := false
	for test := range loader {
		tests = append(tests, test)
		if test.GetStatus() == "focus" {
			hasFocused = true
		}
	}

	client, err := newClient()
	if err != nil {
		return nil, err
	}

	totalTests := 0
	failedTests := 0
	skippedTests := 0
	brokenTests := 0

	for _, v := range tests {
		if hasFocused {
			switch v.GetStatus() {
			case "focus":
				v.SetStatus("")
			case "broken":
				// do nothing
			default:
				v.SetStatus("skipped")
			}
		}

		testResult, err := r.executeTest(v, client)
		switch {
		case err != nil && errors.Is(err, errTestSkipped):
			skippedTests++
		case err != nil && errors.Is(err, errTestBroken):
			brokenTests++
		case err != nil:
			// todo: populate error with test name. Currently it is not possible here to get test name.
			return nil, err
		}

		totalTests++
		if len(testResult.Errors) > 0 {
			failedTests++
		}
		for _, o := range r.output {
			if err := o.Process(v, testResult); err != nil {
				return nil, err
			}
		}
	}

	s := &models.Summary{
		Success: failedTests == 0,
		Skipped: skippedTests,
		Broken:  brokenTests,
		Failed:  failedTests,
		Total:   totalTests,
	}

	return s, nil
}

var (
	errTestSkipped = errors.New("test was skipped")
	errTestBroken  = errors.New("test was broken")
)

func (r *Runner) executeTest(v models.TestInterface, client *http.Client) (*models.Result, error) {

	if v.GetStatus() != "" {
		if v.GetStatus() == "broken" {
			return &models.Result{Test: v}, errTestBroken
		}

		if v.GetStatus() == "skipped" {
			return &models.Result{Test: v}, errTestSkipped
		}
	}

	r.config.Variables.Load(v.GetVariables())
	v = r.config.Variables.Apply(v)

	// load fixtures
	if r.config.FixturesLoader != nil && v.Fixtures() != nil {
		if err := r.config.FixturesLoader.Load(v.Fixtures()); err != nil {
			return nil, fmt.Errorf("unable to load fixtures [%s], error:\n%s", strings.Join(v.Fixtures(), ", "), err)
		}
	}

	// reset mocks
	if r.config.Mocks != nil {
		// prevent deriving the definition from previous test
		r.config.Mocks.ResetDefinitions()
		r.config.Mocks.ResetRunningContext()
	}

	// load mocks
	if r.config.MocksLoader != nil && v.ServiceMocks() != nil {
		if err := r.config.MocksLoader.Load(v.ServiceMocks()); err != nil {
			return nil, err
		}
	}

	// launch script in cmd interface
	if v.BeforeScriptPath() != "" {
		if err := cmd_runner.CmdRun(v.BeforeScriptPath(), v.BeforeScriptTimeout()); err != nil {
			return nil, err
		}
	}

	// make pause
	pause := v.Pause()
	if pause > 0 {
		time.Sleep(time.Duration(pause) * time.Second)
		fmt.Printf("Sleep %ds before requests\n", pause)
	}

	req, err := newRequest(r.config.Host, v)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)

	_ = resp.Body.Close()

	if err != nil {
		return nil, err
	}

	bodyStr := string(body)

	result := models.Result{
		Path:                req.URL.Path,
		Query:               req.URL.RawQuery,
		RequestBody:         actualRequestBody(req),
		ResponseBody:        bodyStr,
		ResponseContentType: resp.Header.Get("Content-Type"),
		ResponseStatusCode:  resp.StatusCode,
		ResponseStatus:      resp.Status,
		ResponseHeaders:     resp.Header,
		Test:                v,
	}

	// launch script in cmd interface
	if v.AfterRequestScriptPath() != "" {
		if err := cmd_runner.CmdRun(v.AfterRequestScriptPath(), v.AfterRequestScriptTimeout()); err != nil {
			return nil, err
		}
	}

	if r.config.Mocks != nil {
		errs := r.config.Mocks.EndRunningContext()
		result.Errors = append(result.Errors, errs...)
	}

	if err := r.setVariablesFromResponse(v, result.ResponseContentType, bodyStr, resp.StatusCode); err != nil {
		return nil, err
	}

	r.config.Variables.Load(v.GetVariables())
	v = r.config.Variables.Apply(v)

	for _, c := range r.checkers {
		errs, err := c.Check(v, &result)
		if err != nil {
			return nil, err
		}
		result.Errors = append(result.Errors, errs...)
	}

	return &result, nil
}

func (r *Runner) setVariablesFromResponse(t models.TestInterface, contentType, body string, statusCode int) error {

	varTemplates := t.GetVariablesToSet()
	if varTemplates == nil {
		return nil
	}

	isJson := strings.Contains(contentType, "json") && body != ""

	vars, err := variables.FromResponse(varTemplates[statusCode], body, isJson)
	if err != nil {
		return err
	}

	if vars == nil {
		return nil
	}

	r.config.Variables.Merge(vars)

	return nil
}
