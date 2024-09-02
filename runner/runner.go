package runner

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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
	HTTPProxyURL   *url.URL
}

type (
	testExecutor func(models.TestInterface) (*models.Result, error)
	testHandler  func(models.TestInterface, testExecutor) error
)

type Runner struct {
	loader               testloader.LoaderInterface
	testExecutionHandler testHandler
	output               []output.OutputInterface
	checkers             []checker.CheckerInterface
	client               *http.Client

	config *Config
}

func New(config *Config, loader testloader.LoaderInterface, handler testHandler) *Runner {
	return &Runner{
		config:               config,
		loader:               loader,
		testExecutionHandler: handler,
		client:               newClient(config.HTTPProxyURL),
	}
}

func (r *Runner) AddOutput(o ...output.OutputInterface) {
	r.output = append(r.output, o...)
}

func (r *Runner) AddCheckers(c ...checker.CheckerInterface) {
	r.checkers = append(r.checkers, c...)
}

func (r *Runner) Run() error {
	tests, err := r.loader.Load()
	if err != nil {
		return err
	}

	hasFocused := checkHasFocused(tests)
	for _, t := range tests {
		// make a copy because go test runner runs tests in separate goroutines
		// and without copy tests will override each other
		test := t
		if hasFocused {
			switch test.GetStatus() {
			case "focus":
				test.SetStatus("")
			case "broken":
				// do nothing
			default:
				test.SetStatus("skipped")
			}
		}

		testExecutor := func(testInterface models.TestInterface) (*models.Result, error) {
			testResult, err := r.executeTest(test)
			if err != nil {
				return nil, err
			}

			for _, o := range r.output {
				if err := o.Process(test, testResult); err != nil {
					return nil, err
				}
			}

			return testResult, nil
		}
		err := r.testExecutionHandler(test, testExecutor)
		if err != nil {
			return fmt.Errorf("test %s error: %s", test.GetName(), err)
		}
	}

	return nil
}

var (
	errTestSkipped = errors.New("test was skipped")
	errTestBroken  = errors.New("test was broken")
)

func (r *Runner) executeTest(v models.TestInterface) (*models.Result, error) {
	if v.GetStatus() != "" {
		if v.GetStatus() == "broken" {
			return &models.Result{Test: v}, errTestBroken
		}

		if v.GetStatus() == "skipped" {
			return &models.Result{Test: v}, errTestSkipped
		}
	}

	r.config.Variables.Load(v.GetCombinedVariables())
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

	resp, err := r.client.Do(req)
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

	r.config.Variables.Load(v.GetCombinedVariables())
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

	isJSON := strings.Contains(contentType, "json") && body != ""

	vars, err := variables.FromResponse(varTemplates[statusCode], body, isJSON)
	if err != nil {
		return err
	}

	if vars == nil {
		return nil
	}

	r.config.Variables.Merge(vars)

	return nil
}

func checkHasFocused(tests []models.TestInterface) bool {
	for _, test := range tests {
		if test.GetStatus() == "focus" {
			return true
		}
	}

	return false
}
