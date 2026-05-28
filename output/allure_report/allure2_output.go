package allure_report

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/lamoda/gonkey/mocks"
	"github.com/lamoda/gonkey/models"
	"github.com/lamoda/gonkey/output/allure_report/allure2"
)

// Allure2Output generates Allure 2 JSON format reports
type Allure2Output struct {
	reportLocation   string
	defaultPackage   string
	defaultTestClass string
}

// NewAllure2Output creates a new Allure 2 output handler
func NewAllure2Output(reportLocation string) *Allure2Output {
	resultsDir, _ := filepath.Abs(reportLocation)
	_ = os.MkdirAll(resultsDir, 0755)

	return &Allure2Output{
		reportLocation: resultsDir,
	}
}

// WithDefaultLabels sets default package and testClass labels
func (o *Allure2Output) WithDefaultLabels(packageLabel, testClassLabel string) *Allure2Output {
	o.defaultPackage = packageLabel
	o.defaultTestClass = testClassLabel
	return o
}

func (o *Allure2Output) Process(t models.TestInterface, result *models.Result) error {
	allureResult := allure2.NewResult(t.GetName(), o.reportLocation)

	if desc := t.GetDescription(); desc != "" {
		allureResult.WithDescription(desc)
	}

	status, err := result.AllureStatus()
	allureResult.WithStatus(status)

	if err != nil {
		allureResult.WithStatusDetails(err.Error(), "")
	}

	hasPackageLabel := false
	hasTestClassLabel := false

	if metadata := t.GetAllureMetadata(); metadata != nil {
		for _, label := range metadata.Labels {
			allureResult.AddLabel(label.Name, label.Value)
			if label.Name == "package" {
				hasPackageLabel = true
			}
			if label.Name == "testClass" {
				hasTestClassLabel = true
			}
		}

		for _, link := range metadata.Links {
			allureResult.AddLink(link.Name, link.URL, link.Type)
		}

		for _, param := range metadata.Parameters {
			allureResult.AddParameter(param.Name, param.Value)
		}
	}

	if !hasPackageLabel && o.defaultPackage != "" {
		allureResult.AddLabel("package", o.defaultPackage)
	}
	if !hasTestClassLabel && o.defaultTestClass != "" {
		allureResult.AddLabel("testClass", o.defaultTestClass)
	}

	allureResult.AddLabel(allure2.LabelFramework, "gonkey")
	allureResult.AddLabel(allure2.LabelLanguage, "go")

	if result.Path != "" {
		allureResult.AddLabel(allure2.LabelStory, result.Path)
	}

	if len(t.GetCombinedVariables()) > 0 {
		for k, v := range t.GetCombinedVariables() {
			allureResult.AddParameter(k, v)
		}
	}

	o.addPreparationStep(allureResult, t)
	if err := o.addRequestStep(allureResult, t, result); err != nil {
		return fmt.Errorf("failed to add request step: %w", err)
	}
	if err := o.addVerificationSteps(allureResult, t, result); err != nil {
		return fmt.Errorf("failed to add verification steps: %w", err)
	}

	allureResult.Finish()
	if err := allureResult.Save(); err != nil {
		return fmt.Errorf("failed to save allure result: %w", err)
	}

	return nil
}

func (o *Allure2Output) addPreparationStep(result *allure2.Result, t models.TestInterface) {
	hasFixtures := len(t.Fixtures()) > 0 || len(t.FixturesMultiDb()) > 0
	hasMocks := len(t.ServiceMocks()) > 0

	if !hasFixtures && !hasMocks {
		return
	}

	prepStep := result.StartStep("Подготовка тестовых данных")

	if hasFixtures {
		fixturesInfo := formatFixturesInfo(t.Fixtures(), t.FixturesMultiDb())
		if fixturesInfo != "" {
			fixtureStep := prepStep.StartSubStep("Загрузка фикстур в БД")
			fixtureStep.AddParameter("files", fixturesInfo)
			fixtureStep.Finish(allure2.StatusPassed)
		}
	}

	if hasMocks {
		mocksInfo := extractMocksInfo(t.ServiceMocks())
		if len(mocksInfo) > 0 {
			mocksStep := prepStep.StartSubStep("Настройка mock-сервисов")

			for _, info := range mocksInfo {
				serviceStep := mocksStep.StartSubStep(formatMockBrief(info))
				serviceStep.Finish(allure2.StatusPassed)
			}

			mocksStep.Finish(allure2.StatusPassed)
		}
	}

	prepStep.Finish(allure2.StatusPassed)
}

func (o *Allure2Output) addRequestStep(result *allure2.Result, t models.TestInterface, testResult *models.Result) error {
	stepName := fmt.Sprintf("Отправка %s запроса к %s", t.GetMethod(), testResult.Path)
	requestStep := result.StartStep(stepName)

	requestStep.AddParameter("method", t.GetMethod())
	if testResult.Query != "" {
		requestStep.AddParameter("query", testResult.Query)
	}

	if testResult.RequestBody != "" {
		if err := requestStep.AddAttachment("Request Body", testResult.RequestBody,
			allure2.MimeTypeApplicationJSON, o.reportLocation); err != nil {
			return err
		}
	}

	if len(t.Headers()) > 0 {
		headersContent := formatHeaders(t.Headers())
		if err := requestStep.AddAttachment("Request Headers", headersContent,
			allure2.MimeTypeTextPlain, o.reportLocation); err != nil {
			return err
		}
	}

	if len(t.Cookies()) > 0 {
		cookiesContent := formatCookies(t.Cookies())
		if err := requestStep.AddAttachment("Request Cookies", cookiesContent,
			allure2.MimeTypeTextPlain, o.reportLocation); err != nil {
			return err
		}
	}

	requestStep.Finish(allure2.StatusPassed)
	return nil
}

func (o *Allure2Output) addVerificationSteps(result *allure2.Result, t models.TestInterface, testResult *models.Result) error {
	errorCategories := categorizeErrors(testResult.Errors)

	if err := o.addResponseVerificationStep(result, t, testResult, errorCategories); err != nil {
		return err
	}

	if err := o.addDatabaseVerificationStep(result, testResult, errorCategories); err != nil {
		return err
	}

	if err := o.addMockVerificationStep(result, t, errorCategories); err != nil {
		return err
	}

	return nil
}

func (o *Allure2Output) addResponseVerificationStep(
	result *allure2.Result,
	t models.TestInterface,
	testResult *models.Result,
	errorCategories map[models.ErrorCategory]ErrorsByIdentifier,
) error {
	hasStatusCodeError := len(errorCategories[models.ErrorCategoryStatusCode]) > 0
	hasBodyError := len(errorCategories[models.ErrorCategoryResponseBody]) > 0
	hasHeaderError := len(errorCategories[models.ErrorCategoryResponseHeader]) > 0

	responseStepStatus := allure2.StatusPassed
	if hasStatusCodeError || hasBodyError || hasHeaderError {
		responseStepStatus = allure2.StatusFailed
	}

	responseStep := result.StartStep("Проверка ответа сервера")

	statusCodeStatus := allure2.StatusPassed
	if hasStatusCodeError {
		statusCodeStatus = allure2.StatusFailed
	}

	expectedCodes := getExpectedStatusCodes(t.GetResponses())
	statusStepName := "Проверка статус кода"
	if len(expectedCodes) > 0 {
		statusStepName = fmt.Sprintf("Проверка статус кода (ожидается: %s)", formatStatusCodes(expectedCodes))
	}

	statusStep := responseStep.StartSubStep(statusStepName)
	statusStep.AddParameter("actual", fmt.Sprintf("%d", testResult.ResponseStatusCode))
	statusStep.Finish(statusCodeStatus)

	bodyStepStatus := allure2.StatusPassed
	if hasBodyError {
		bodyStepStatus = allure2.StatusFailed
	}

	bodyStep := responseStep.StartSubStep("Проверка тела ответа")
	if testResult.ResponseBody != "" {
		if err := bodyStep.AddAttachment("Response Body", testResult.ResponseBody,
			allure2.MimeTypeApplicationJSON, o.reportLocation); err != nil {
			return err
		}
	}
	bodyStep.Finish(bodyStepStatus)

	expectedHeaders, hasExpectedHeaders := t.GetResponseHeaders(testResult.ResponseStatusCode)
	if hasExpectedHeaders && len(expectedHeaders) > 0 {
		headerStepStatus := allure2.StatusPassed
		if hasHeaderError {
			headerStepStatus = allure2.StatusFailed
		}

		headerStep := responseStep.StartSubStep("Проверка заголовков ответа")
		if len(testResult.ResponseHeaders) > 0 {
			headersContent := formatResponseHeaders(testResult.ResponseHeaders)
			if err := headerStep.AddAttachment("Response Headers", headersContent,
				allure2.MimeTypeTextPlain, o.reportLocation); err != nil {
				return err
			}
		}
		headerStep.Finish(headerStepStatus)
	}

	responseStep.Finish(responseStepStatus)
	return nil
}

func (o *Allure2Output) addDatabaseVerificationStep(
	result *allure2.Result,
	testResult *models.Result,
	errorCategories map[models.ErrorCategory]ErrorsByIdentifier,
) error {
	if len(testResult.DatabaseResult) == 0 {
		return nil
	}

	dbErrors := errorCategories[models.ErrorCategoryDatabase]
	dbStep := result.StartStep("Проверка данных в БД")

	for i, dbResult := range testResult.DatabaseResult {
		if dbResult.Query != "" {
			queryIdentifier := fmt.Sprintf("%d", i)
			queryHasError := len(dbErrors[queryIdentifier]) > 0
			queryStepStatus := allure2.StatusPassed
			if queryHasError {
				queryStepStatus = allure2.StatusFailed
			}

			dbQueryStep := dbStep.StartSubStep(fmt.Sprintf("DB Query #%d", i+1))

			queryContent := fmt.Sprintf("SQL: %s", dbResult.Query)
			if err := dbQueryStep.AddAttachment("Query", queryContent,
				allure2.MimeTypeTextPlain, o.reportLocation); err != nil {
				return err
			}

			if len(dbResult.Response) > 0 {
				responseContent := formatDbResponse(dbResult.Response)
				if err := dbQueryStep.AddAttachment("Response", responseContent,
					allure2.MimeTypeTextPlain, o.reportLocation); err != nil {
					return err
				}
			}

			dbQueryStep.Finish(queryStepStatus)
		}
	}

	dbStepStatus := allure2.StatusPassed
	if len(dbErrors) > 0 {
		dbStepStatus = allure2.StatusFailed
	}
	dbStep.Finish(dbStepStatus)
	return nil
}

func (o *Allure2Output) addMockVerificationStep(
	result *allure2.Result,
	t models.TestInterface,
	errorCategories map[models.ErrorCategory]ErrorsByIdentifier,
) error {
	if len(t.ServiceMocks()) == 0 {
		return nil
	}

	mockErrors := errorCategories[models.ErrorCategoryMock]
	mockStep := result.StartStep("Проверка вызовов mock-сервисов")

	mocksInfo := extractMocksInfo(t.ServiceMocks())
	for _, info := range mocksInfo {
		serviceErrors := mockErrors[info.ServiceName]
		o.addMockVerificationSubsteps(mockStep, info, serviceErrors)
	}

	mockStepStatus := allure2.StatusPassed
	if len(mockErrors) > 0 {
		mockStepStatus = allure2.StatusFailed
	}
	mockStep.Finish(mockStepStatus)
	return nil
}

func formatHeaders(headers map[string]string) string {
	var lines []string
	for k, v := range headers {
		lines = append(lines, fmt.Sprintf("%s: %s", k, v))
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

func formatResponseHeaders(headers map[string][]string) string {
	var lines []string
	for k, values := range headers {
		for _, v := range values {
			lines = append(lines, fmt.Sprintf("%s: %s", k, v))
		}
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

func formatCookies(cookies map[string]string) string {
	var lines []string
	for k, v := range cookies {
		lines = append(lines, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(lines, "\n")
}

func formatDbResponse(response []string) string {
	if len(response) == 0 {
		return "[]"
	}
	if len(response) > 10 {
		return fmt.Sprintf("[\n  %s,\n  ... (%d total rows)\n]",
			response[0], len(response))
	}
	return fmt.Sprintf("[\n  %s\n]", strings.Join(response, ",\n  "))
}

func getExpectedStatusCodes(responses map[int]string) []int {
	codes := make([]int, 0, len(responses))
	for code := range responses {
		codes = append(codes, code)
	}
	sort.Ints(codes)
	return codes
}

func formatStatusCodes(codes []int) string {
	if len(codes) == 0 {
		return ""
	}
	if len(codes) == 1 {
		return fmt.Sprintf("%d", codes[0])
	}
	strCodes := make([]string, len(codes))
	for i, code := range codes {
		strCodes[i] = fmt.Sprintf("%d", code)
	}
	return strings.Join(strCodes, ", ")
}

type ErrorsByIdentifier map[string][]error

func categorizeErrors(errs []error) map[models.ErrorCategory]ErrorsByIdentifier {
	categories := make(map[models.ErrorCategory]ErrorsByIdentifier)

	for _, err := range errs {
		var checkErr *models.CheckError
		var mockErr *mocks.Error

		if errors.As(err, &checkErr) {
			category := checkErr.GetCategory()
			identifier := checkErr.GetIdentifier()

			if categories[category] == nil {
				categories[category] = make(ErrorsByIdentifier)
			}
			categories[category][identifier] = append(categories[category][identifier], err)
		} else if errors.As(err, &mockErr) {
			if categories[models.ErrorCategoryMock] == nil {
				categories[models.ErrorCategoryMock] = make(ErrorsByIdentifier)
			}
			categories[models.ErrorCategoryMock][mockErr.ServiceName] = append(
				categories[models.ErrorCategoryMock][mockErr.ServiceName], err)
		} else {
			// Uncategorized errors are treated as body errors for backward compatibility
			if categories[models.ErrorCategoryResponseBody] == nil {
				categories[models.ErrorCategoryResponseBody] = make(ErrorsByIdentifier)
			}
			categories[models.ErrorCategoryResponseBody][""] = append(categories[models.ErrorCategoryResponseBody][""], err)
		}
	}

	return categories
}

func (o *Allure2Output) Finalize() {
}

func groupErrorsByEndpoint(errs []error) map[string][]error {
	grouped := make(map[string][]error)

	for _, err := range errs {
		var constraintErr *mocks.RequestConstraintError
		var callsErr *mocks.CallsMismatchError

		if errors.As(err, &constraintErr) && constraintErr.Endpoint != "" {
			endpoint := extractURIFromPath(constraintErr.Endpoint)
			grouped[endpoint] = append(grouped[endpoint], err)
		} else if errors.As(err, &callsErr) {
			endpoint := extractURIFromPath(callsErr.Path)
			grouped[endpoint] = append(grouped[endpoint], err)
		} else {
			grouped[""] = append(grouped[""], err)
		}
	}

	return grouped
}

func extractURIFromPath(path string) string {
	prefixes := []string{"$.uriVary.", "$.methodVary."}
	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			return path[len(prefix):]
		}
	}

	if strings.HasPrefix(path, "$.basedOnRequest.") {
		num := path[len("$.basedOnRequest."):]
		if _, err := strconv.Atoi(num); err == nil {
			return fmt.Sprintf("case %s", num)
		}
	}

	if strings.HasPrefix(path, "$.sequence.") {
		num := path[len("$.sequence."):]
		if _, err := strconv.Atoi(num); err == nil {
			return fmt.Sprintf("step %s", num)
		}
	}

	if strings.HasPrefix(path, "$.") && len(path) > 2 {
		num := path[2:]
		if _, err := strconv.Atoi(num); err == nil {
			return fmt.Sprintf("step %s", num)
		}
	}

	if idx := strings.Index(path, "/"); idx >= 0 {
		return path[idx:]
	}
	return ""
}

func formatConstraintError(err error) string {
	var constraintErr *mocks.RequestConstraintError
	var callsErr *mocks.CallsMismatchError

	if errors.As(err, &constraintErr) {
		kind := reflect.TypeOf(constraintErr.Constraint).String()
		kind = strings.TrimPrefix(kind, "*mocks.")
		kind = strings.TrimSuffix(kind, "Constraint")
		kind = strings.TrimSuffix(kind, "constraint")
		return fmt.Sprintf("%s: %s", kind, constraintErr.Unwrap().Error())
	}
	if errors.As(err, &callsErr) {
		return fmt.Sprintf("calls: expected %d, actual %d", callsErr.Expected, callsErr.Actual)
	}
	return err.Error()
}

func (o *Allure2Output) addMockVerificationSubsteps(
	mockStep *allure2.Step,
	info mockInfo,
	serviceErrors []error,
) {
	hasErrors := len(serviceErrors) > 0

	if !hasErrors {
		serviceStep := mockStep.StartSubStep(info.ServiceName)
		serviceStep.Finish(allure2.StatusPassed)
		return
	}

	serviceStep := mockStep.StartSubStep(fmt.Sprintf("%s: FAILED", info.ServiceName))

	switch info.Strategy {
	case "uriVary", "methodVary":
		o.addEndpointGroupedErrors(serviceStep, info, serviceErrors)

	case "sequence", "basedOnRequest":
		o.addStrategyGroupedErrors(serviceStep, info, serviceErrors)

	default:
		for _, err := range serviceErrors {
			errStep := serviceStep.StartSubStep(formatConstraintError(err))
			errStep.Finish(allure2.StatusFailed)
		}
	}

	serviceStep.Finish(allure2.StatusFailed)
}

func (o *Allure2Output) addEndpointGroupedErrors(
	serviceStep *allure2.Step,
	info mockInfo,
	serviceErrors []error,
) {
	errorsByEndpoint := groupErrorsByEndpoint(serviceErrors)

	for _, endpoint := range info.Endpoints {
		if endpointErrors := errorsByEndpoint[endpoint]; len(endpointErrors) > 0 {
			endpointStep := serviceStep.StartSubStep(fmt.Sprintf("%s: FAILED", endpoint))
			for _, err := range endpointErrors {
				errStep := endpointStep.StartSubStep(formatConstraintError(err))
				errStep.Finish(allure2.StatusFailed)
			}
			endpointStep.Finish(allure2.StatusFailed)
		}
	}

	if ungroupedErrors := errorsByEndpoint[""]; len(ungroupedErrors) > 0 {
		for _, err := range ungroupedErrors {
			errStep := serviceStep.StartSubStep(formatConstraintError(err))
			errStep.Finish(allure2.StatusFailed)
		}
	}
}

func (o *Allure2Output) addStrategyGroupedErrors(
	serviceStep *allure2.Step,
	info mockInfo,
	serviceErrors []error,
) {
	errorsByEndpoint := groupErrorsByEndpoint(serviceErrors)

	strategyStep := serviceStep.StartSubStep(info.Strategy)

	hasStepErrors := false
	for _, endpoint := range info.Endpoints {
		endpointErrors := errorsByEndpoint[endpoint]
		if len(endpointErrors) > 0 {
			hasStepErrors = true
			endpointStep := strategyStep.StartSubStep(fmt.Sprintf("%s: FAILED", endpoint))
			for _, err := range endpointErrors {
				errStep := endpointStep.StartSubStep(formatConstraintError(err))
				errStep.Finish(allure2.StatusFailed)
			}
			endpointStep.Finish(allure2.StatusFailed)
		}
	}

	if ungroupedErrors := errorsByEndpoint[""]; len(ungroupedErrors) > 0 {
		for _, err := range ungroupedErrors {
			errStep := strategyStep.StartSubStep(formatConstraintError(err))
			errStep.Finish(allure2.StatusFailed)
		}
		hasStepErrors = true
	}

	if hasStepErrors {
		strategyStep.Finish(allure2.StatusFailed)
	} else {
		strategyStep.Finish(allure2.StatusPassed)
	}
}
