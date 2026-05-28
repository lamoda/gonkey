package allure2

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResult(t *testing.T) {
	result := NewResult("Test Name", "/tmp")

	assert.NotEmpty(t, result.UUID, "UUID should be generated")
	assert.Equal(t, "Test Name", result.Name)
	assert.Equal(t, "Test Name", result.FullName)
	assert.NotEmpty(t, result.HistoryID, "HistoryID should be generated")
	assert.Equal(t, StatusUnknown, result.Status)
	assert.Equal(t, StageRunning, result.Stage)
	assert.Greater(t, result.Start, int64(0), "Start timestamp should be set")
	assert.Equal(t, int64(0), result.Stop, "Stop should not be set initially")
}

func TestResultFinish(t *testing.T) {
	result := NewResult("Test", "/tmp")
	initialStart := result.Start

	result.Finish()

	assert.Greater(t, result.Stop, int64(0), "Stop timestamp should be set")
	assert.GreaterOrEqual(t, result.Stop, initialStart, "Stop should be >= Start")
	assert.Equal(t, StageFinished, result.Stage)
}

func TestResultWithMethods(t *testing.T) {
	result := NewResult("Test", "/tmp")

	result.
		WithDescription("Test description").
		WithStatus(StatusPassed).
		AddLabel(LabelSeverity, "critical").
		AddLabel(LabelFeature, "authentication").
		AddLink("User Story", "https://jira.example.com/US-123", LinkTypeIssue).
		AddParameter("environment", "staging")

	assert.Equal(t, "Test description", result.Description)
	assert.Equal(t, StatusPassed, result.Status)
	assert.Len(t, result.Labels, 2)
	assert.Len(t, result.Links, 1)
	assert.Len(t, result.Parameters, 1)
}

func TestResultAddAttachment(t *testing.T) {
	tmpDir := t.TempDir()

	result := NewResult("Test", tmpDir)
	content := "Request body content"

	err := result.AddAttachment("Request", content, MimeTypeTextPlain)
	require.NoError(t, err)

	require.Len(t, result.Attachments, 1)
	attachment := result.Attachments[0]
	assert.Equal(t, "Request", attachment.Name)
	assert.Equal(t, MimeTypeTextPlain, attachment.Type)
	assert.NotEmpty(t, attachment.Source)

	attachmentPath := filepath.Join(tmpDir, attachment.Source)
	data, err := os.ReadFile(attachmentPath)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))
}

func TestResultSave(t *testing.T) {
	tmpDir := t.TempDir()

	result := NewResult("Test Save", tmpDir)
	result.
		WithDescription("Test description").
		WithStatus(StatusPassed).
		AddLabel(LabelSeverity, "critical")

	err := result.Save()
	require.NoError(t, err)

	resultFile := filepath.Join(tmpDir, result.UUID+"-result.json")
	assert.FileExists(t, resultFile)

	data, err := os.ReadFile(resultFile)
	require.NoError(t, err)

	var loaded Result
	err = json.Unmarshal(data, &loaded)
	require.NoError(t, err)

	assert.Equal(t, result.UUID, loaded.UUID)
	assert.Equal(t, result.Name, loaded.Name)
	assert.Equal(t, result.Status, loaded.Status)
	assert.Equal(t, result.Description, loaded.Description)
	assert.Greater(t, loaded.Stop, int64(0), "Stop should be set after Save")
	assert.Len(t, loaded.Labels, 1)
}

func TestResultJSONStructure(t *testing.T) {
	result := NewResult("JSON Test", "/tmp")
	result.
		WithDescription("Test JSON structure").
		WithStatus(StatusFailed).
		WithStatusDetails("Assertion failed", "stack trace here").
		AddLabel(LabelSeverity, "blocker").
		AddLabel(LabelFeature, "api").
		AddLink("Bug", "https://jira.example.com/BUG-789", LinkTypeIssue).
		AddParameter("version", "1.0.0")

	result.Attachments = []Attachment{
		{Name: "Request", Source: "req-123.txt", Type: MimeTypeTextPlain},
	}
	result.Finish()

	data, err := json.MarshalIndent(result, "", "  ")
	require.NoError(t, err)

	var jsonMap map[string]interface{}
	err = json.Unmarshal(data, &jsonMap)
	require.NoError(t, err)

	// Allure 2 required fields
	assert.Contains(t, jsonMap, "uuid")
	assert.Contains(t, jsonMap, "name")
	assert.Contains(t, jsonMap, "status")
	assert.Contains(t, jsonMap, "start")
	assert.Contains(t, jsonMap, "stop")
	assert.Contains(t, jsonMap, "labels")
	assert.Contains(t, jsonMap, "links")
	assert.Contains(t, jsonMap, "parameters")
	assert.Contains(t, jsonMap, "attachments")
	assert.Contains(t, jsonMap, "statusDetails")

	statusDetails := jsonMap["statusDetails"].(map[string]interface{})
	assert.Equal(t, "Assertion failed", statusDetails["message"])
	assert.Equal(t, "stack trace here", statusDetails["trace"])
}

func TestResultOmitEmptyFields(t *testing.T) {
	result := NewResult("Minimal Test", "/tmp")
	result.WithStatus(StatusPassed)
	result.Finish()

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var jsonMap map[string]interface{}
	err = json.Unmarshal(data, &jsonMap)
	require.NoError(t, err)

	assert.Contains(t, jsonMap, "uuid")
	assert.Contains(t, jsonMap, "name")
	assert.Contains(t, jsonMap, "status")
	assert.Contains(t, jsonMap, "start")
	assert.Contains(t, jsonMap, "stop")

	// Empty fields should be omitted
	_, hasDescription := jsonMap["description"]
	assert.False(t, hasDescription, "Empty description should be omitted")

	_, hasStatusDetails := jsonMap["statusDetails"]
	assert.False(t, hasStatusDetails, "Nil statusDetails should be omitted")
}

func TestLabelHelpers(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(string) Label
		value    string
		expected Label
	}{
		{"Severity", NewSeverityLabel, "critical", Label{Name: LabelSeverity, Value: "critical"}},
		{"Feature", NewFeatureLabel, "auth", Label{Name: LabelFeature, Value: "auth"}},
		{"Story", NewStoryLabel, "login", Label{Name: LabelStory, Value: "login"}},
		{"Epic", NewEpicLabel, "security", Label{Name: LabelEpic, Value: "security"}},
		{"Owner", NewOwnerLabel, "team-api", Label{Name: LabelOwner, Value: "team-api"}},
		{"Tag", NewTagLabel, "smoke", Label{Name: LabelTag, Value: "smoke"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label := tt.fn(tt.value)
			assert.Equal(t, tt.expected, label)
		})
	}
}

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		mimeType string
		expected string
	}{
		{MimeTypeTextPlain, "txt"},
		{MimeTypeTextHTML, "html"},
		{MimeTypeApplicationJSON, "json"},
		{MimeTypeApplicationXML, "xml"},
		{MimeTypeImagePNG, "png"},
		{MimeTypeImageJPEG, "jpg"},
		{"application/octet-stream", "txt"}, // default
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			ext := getFileExtension(tt.mimeType)
			assert.Equal(t, tt.expected, ext)
		})
	}
}

func TestGenerateHistoryID(t *testing.T) {
	// Same name should generate same history ID
	id1 := generateHistoryID("Test Name")
	id2 := generateHistoryID("Test Name")
	assert.Equal(t, id1, id2, "Same name should produce same history ID")

	// Different names should generate different history IDs
	id3 := generateHistoryID("Different Test")
	assert.NotEqual(t, id1, id3, "Different names should produce different history IDs")

	// History ID should be a valid hex string (MD5)
	assert.Len(t, id1, 32, "MD5 hash should be 32 characters")
}

func TestResultWithStatusDetails(t *testing.T) {
	result := NewResult("Test", "/tmp")
	result.WithStatusDetails("Error message", "Full stack trace")

	require.NotNil(t, result.StatusDetails)
	assert.Equal(t, "Error message", result.StatusDetails.Message)
	assert.Equal(t, "Full stack trace", result.StatusDetails.Trace)
}

func TestAddLabels(t *testing.T) {
	result := NewResult("Test", "/tmp")

	labels := []Label{
		NewSeverityLabel("critical"),
		NewFeatureLabel("api"),
		NewEpicLabel("authentication"),
	}

	result.AddLabels(labels...)

	assert.Len(t, result.Labels, 3)
	assert.Equal(t, LabelSeverity, result.Labels[0].Name)
	assert.Equal(t, LabelFeature, result.Labels[1].Name)
	assert.Equal(t, LabelEpic, result.Labels[2].Name)
}

func TestMultipleAttachments(t *testing.T) {
	tmpDir := t.TempDir()
	result := NewResult("Test", tmpDir)

	err := result.AddAttachment("Request", "POST /api/users", MimeTypeTextPlain)
	require.NoError(t, err)

	err = result.AddAttachment("Response", `{"id": 123}`, MimeTypeApplicationJSON)
	require.NoError(t, err)

	err = result.AddAttachment("SQL Query", "SELECT * FROM users", MimeTypeTextPlain)
	require.NoError(t, err)

	assert.Len(t, result.Attachments, 3)

	for _, attachment := range result.Attachments {
		attachmentPath := filepath.Join(tmpDir, attachment.Source)
		assert.FileExists(t, attachmentPath)
	}
}

func TestFormatContentIfJSON(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		mimeType string
		expected string
	}{
		{
			name:     "valid compact JSON should be formatted",
			content:  `{"id":123,"name":"test","nested":{"key":"value"}}`,
			mimeType: MimeTypeApplicationJSON,
			expected: "{\n  \"id\": 123,\n  \"name\": \"test\",\n  \"nested\": {\n    \"key\": \"value\"\n  }\n}",
		},
		{
			name:     "invalid JSON should return as-is",
			content:  `{"invalid": json}`,
			mimeType: MimeTypeApplicationJSON,
			expected: `{"invalid": json}`,
		},
		{
			name:     "JSON array should be formatted",
			content:  `[{"id":1},{"id":2}]`,
			mimeType: MimeTypeApplicationJSON,
			expected: "[\n  {\n    \"id\": 1\n  },\n  {\n    \"id\": 2\n  }\n]",
		},
		{
			name:     "non-JSON mime type should return as-is",
			content:  `{"id":123}`,
			mimeType: MimeTypeTextPlain,
			expected: `{"id":123}`,
		},
		{
			name:     "empty JSON object should be formatted",
			content:  `{}`,
			mimeType: MimeTypeApplicationJSON,
			expected: "{}",
		},
		{
			name:     "empty JSON array should be formatted",
			content:  `[]`,
			mimeType: MimeTypeApplicationJSON,
			expected: "[]",
		},
		{
			name:     "JSON with unicode should be formatted",
			content:  `{"name":"тест","emoji":"😀"}`,
			mimeType: MimeTypeApplicationJSON,
			expected: "{\n  \"emoji\": \"😀\",\n  \"name\": \"тест\"\n}",
		},
		{
			name:     "already formatted JSON should be re-formatted consistently",
			content:  "{\n  \"id\": 123\n}",
			mimeType: MimeTypeApplicationJSON,
			expected: "{\n  \"id\": 123\n}",
		},
		{
			name:     "JSON with special characters should be formatted",
			content:  `{"path":"C:\\Users\\test","url":"https://example.com/api?q=1&p=2"}`,
			mimeType: MimeTypeApplicationJSON,
			expected: "{\n  \"path\": \"C:\\\\Users\\\\test\",\n  \"url\": \"https://example.com/api?q=1\\u0026p=2\"\n}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatContentIfJSON(tt.content, tt.mimeType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAddAttachment_JSONFormatting(t *testing.T) {
	tmpDir := t.TempDir()
	result := NewResult("Test", tmpDir)

	compactJSON := `{"id":123,"name":"test","data":{"nested":"value"}}`

	err := result.AddAttachment("Response", compactJSON, MimeTypeApplicationJSON)
	require.NoError(t, err)

	attachmentPath := filepath.Join(tmpDir, result.Attachments[0].Source)
	content, err := os.ReadFile(attachmentPath)
	require.NoError(t, err)

	var data interface{}
	err = json.Unmarshal(content, &data)
	require.NoError(t, err, "Attachment content should be valid JSON")

	contentStr := string(content)
	assert.Contains(t, contentStr, "\n", "Formatted JSON should contain newlines")
	assert.Contains(t, contentStr, "  ", "Formatted JSON should contain indentation")
}

func TestStepAddAttachment_JSONFormatting(t *testing.T) {
	tmpDir := t.TempDir()
	result := NewResult("Test", tmpDir)

	step := result.StartStep("Test Step")

	compactJSON := `{"status":"ok","items":[1,2,3]}`

	err := step.AddAttachment("Data", compactJSON, MimeTypeApplicationJSON, tmpDir)
	require.NoError(t, err)

	attachmentPath := filepath.Join(tmpDir, step.Attachments[0].Source)
	content, err := os.ReadFile(attachmentPath)
	require.NoError(t, err)

	var data interface{}
	err = json.Unmarshal(content, &data)
	require.NoError(t, err, "Attachment content should be valid JSON")

	contentStr := string(content)
	assert.Contains(t, contentStr, "\n", "Formatted JSON should contain newlines")
	assert.Contains(t, contentStr, "  ", "Formatted JSON should contain indentation")
}

func TestAddAttachment_InvalidJSONPreserved(t *testing.T) {
	tmpDir := t.TempDir()
	result := NewResult("Test", tmpDir)

	invalidJSON := `{"invalid": json}`

	err := result.AddAttachment("Invalid Response", invalidJSON, MimeTypeApplicationJSON)
	require.NoError(t, err)

	attachmentPath := filepath.Join(tmpDir, result.Attachments[0].Source)
	content, err := os.ReadFile(attachmentPath)
	require.NoError(t, err)

	assert.Equal(t, invalidJSON, string(content), "Invalid JSON should be preserved unchanged")
}

func TestStepAddAttachment_InvalidJSONPreserved(t *testing.T) {
	tmpDir := t.TempDir()
	result := NewResult("Test", tmpDir)

	step := result.StartStep("Test Step")
	invalidJSON := `{"broken": json content}`

	err := step.AddAttachment("Invalid Data", invalidJSON, MimeTypeApplicationJSON, tmpDir)
	require.NoError(t, err)

	attachmentPath := filepath.Join(tmpDir, step.Attachments[0].Source)
	content, err := os.ReadFile(attachmentPath)
	require.NoError(t, err)

	assert.Equal(t, invalidJSON, string(content), "Invalid JSON should be preserved unchanged")
}

func TestAddAttachment_NonJSONMimeTypeNotFormatted(t *testing.T) {
	tmpDir := t.TempDir()
	result := NewResult("Test", tmpDir)

	jsonLikeContent := `{"id":123,"name":"test"}`

	err := result.AddAttachment("Plain Text", jsonLikeContent, MimeTypeTextPlain)
	require.NoError(t, err)

	attachmentPath := filepath.Join(tmpDir, result.Attachments[0].Source)
	content, err := os.ReadFile(attachmentPath)
	require.NoError(t, err)

	assert.Equal(t, jsonLikeContent, string(content), "Non-JSON mime type should not be formatted")
	assert.NotContains(t, string(content), "\n", "Should not contain newlines")
}
