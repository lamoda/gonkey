package runner

import (
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

	client, err := newClient()
	if err != nil {
		return nil, err
	}

	totalTests := 0
	failedTests := 0

	for v := range loader {
		testResult, err := r.executeTest(v, client)
		if err != nil {
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
		Failed:  failedTests,
		Total:   totalTests,
	}

	return s, nil
}

func (r *Runner) executeTest(v models.TestInterface, client *http.Client) (*models.Result, error) {

	r.config.Variables.Load(v.GetVariables())
	v = r.config.Variables.Apply(v)

	// load fixtures
	if r.config.FixturesLoader != nil && v.Fixtures() != nil {
		if err := r.config.FixturesLoader.Load(v.Fixtures()); err != nil {
			return nil, err
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
	if err != nil {
		return nil, err
	}
	_ = resp.Body.Close()

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

	if r.config.Mocks != nil {
		errs := r.config.Mocks.EndRunningContext()
		result.Errors = append(result.Errors, errs...)
	}

	for _, c := range r.checkers {
		errs, err := c.Check(v, &result)
		if err != nil {
			return nil, err
		}
		result.Errors = append(result.Errors, errs...)
	}

	if err := r.setVariablesFromResponse(v, result.ResponseContentType, bodyStr, resp.StatusCode); err != nil {
		return nil, err
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
