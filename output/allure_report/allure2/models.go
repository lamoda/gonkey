package allure2

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type Result struct {
	UUID          string         `json:"uuid"`
	HistoryID     string         `json:"historyId,omitempty"`
	FullName      string         `json:"fullName,omitempty"`
	Name          string         `json:"name"`
	Status        string         `json:"status"`
	StatusDetails *StatusDetails `json:"statusDetails,omitempty"`
	Stage         string         `json:"stage,omitempty"`
	Description   string         `json:"description,omitempty"`
	Start         int64          `json:"start"`
	Stop          int64          `json:"stop"`
	Labels        []Label        `json:"labels,omitempty"`
	Links         []Link         `json:"links,omitempty"`
	Parameters    []Parameter    `json:"parameters,omitempty"`
	Attachments   []Attachment   `json:"attachments,omitempty"`
	Steps         []Step         `json:"steps,omitempty"`

	targetDir string `json:"-"`
}

type StatusDetails struct {
	Message string `json:"message,omitempty"`
	Trace   string `json:"trace,omitempty"`
}

type Label struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Link struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url"`
	Type string `json:"type,omitempty"`
}

type Parameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Attachment struct {
	Name   string `json:"name"`
	Source string `json:"source"`
	Type   string `json:"type"`
}

type Step struct {
	Name        string       `json:"name"`
	Status      string       `json:"status"`
	Stage       string       `json:"stage,omitempty"`
	Start       int64        `json:"start"`
	Stop        int64        `json:"stop"`
	Parameters  []Parameter  `json:"parameters,omitempty"`
	Steps       []Step       `json:"steps,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

const (
	StatusPassed  = "passed"
	StatusFailed  = "failed"
	StatusBroken  = "broken"
	StatusSkipped = "skipped"
	StatusUnknown = "unknown"
)

const (
	StageRunning  = "running"
	StageFinished = "finished"
)

const (
	LinkTypeTMS   = "tms"
	LinkTypeIssue = "issue"
	LinkTypeLink  = "link"
)

const (
	LabelSeverity  = "severity"
	LabelStory     = "story"
	LabelFeature   = "feature"
	LabelEpic      = "epic"
	LabelOwner     = "owner"
	LabelTag       = "tag"
	LabelFramework = "framework"
	LabelLanguage  = "language"
)

const (
	MimeTypeTextPlain       = "text/plain"
	MimeTypeTextHTML        = "text/html"
	MimeTypeApplicationJSON = "application/json"
	MimeTypeApplicationXML  = "application/xml"
	MimeTypeImagePNG        = "image/png"
	MimeTypeImageJPEG       = "image/jpeg"
)

func NewResult(name, targetDir string) *Result {
	u := uuid.New().String()
	return &Result{
		UUID:      u,
		Name:      name,
		FullName:  name,
		HistoryID: generateHistoryID(name),
		Start:     currentTimeMillis(),
		Status:    StatusUnknown,
		Stage:     StageRunning,
		targetDir: targetDir,
	}
}

func (r *Result) Finish() {
	r.Stop = currentTimeMillis()
	r.Stage = StageFinished
}

func (r *Result) WithDescription(desc string) *Result {
	r.Description = desc
	return r
}

func (r *Result) WithStatus(status string) *Result {
	r.Status = status
	return r
}

func (r *Result) WithStatusDetails(message, trace string) *Result {
	r.StatusDetails = &StatusDetails{
		Message: message,
		Trace:   trace,
	}
	return r
}

func (r *Result) AddLabel(name, value string) *Result {
	r.Labels = append(r.Labels, Label{
		Name:  name,
		Value: value,
	})
	return r
}

func (r *Result) AddLabels(labels ...Label) *Result {
	r.Labels = append(r.Labels, labels...)
	return r
}

func (r *Result) AddLink(name, url, linkType string) *Result {
	r.Links = append(r.Links, Link{
		Name: name,
		URL:  url,
		Type: linkType,
	})
	return r
}

func (r *Result) AddParameter(name, value string) *Result {
	r.Parameters = append(r.Parameters, Parameter{
		Name:  name,
		Value: value,
	})
	return r
}

// StartStep creates a new step and returns it for further configuration
func (r *Result) StartStep(name string) *Step {
	step := &Step{
		Name:   name,
		Status: StatusUnknown,
		Stage:  StageRunning,
		Start:  currentTimeMillis(),
	}
	r.Steps = append(r.Steps, *step)
	// Return pointer to the last step in the slice
	return &r.Steps[len(r.Steps)-1]
}

// Step methods for nested configuration
func (s *Step) Finish(status string) *Step {
	s.Stop = currentTimeMillis()
	s.Status = status
	s.Stage = StageFinished
	return s
}

func (s *Step) AddParameter(name, value string) *Step {
	s.Parameters = append(s.Parameters, Parameter{
		Name:  name,
		Value: value,
	})
	return s
}

func (s *Step) AddAttachment(name, content, mimeType string, targetDir string) error {
	attachUUID := uuid.New().String()
	ext := getFileExtension(mimeType)
	source := fmt.Sprintf("%s-attachment.%s", attachUUID, ext)

	// Format JSON content if applicable
	formattedContent := formatContentIfJSON(content, mimeType)

	attachmentPath := filepath.Join(targetDir, source)
	if err := os.WriteFile(attachmentPath, []byte(formattedContent), 0644); err != nil {
		return fmt.Errorf("failed to write attachment: %w", err)
	}

	s.Attachments = append(s.Attachments, Attachment{
		Name:   name,
		Source: source,
		Type:   mimeType,
	})

	return nil
}

// StartSubStep creates a nested step within this step
func (s *Step) StartSubStep(name string) *Step {
	step := &Step{
		Name:   name,
		Status: StatusUnknown,
		Stage:  StageRunning,
		Start:  currentTimeMillis(),
	}
	s.Steps = append(s.Steps, *step)
	// Return pointer to the last step in the slice
	return &s.Steps[len(s.Steps)-1]
}

func (r *Result) AddAttachment(name, content, mimeType string) error {
	attachUUID := uuid.New().String()
	ext := getFileExtension(mimeType)
	source := fmt.Sprintf("%s-attachment.%s", attachUUID, ext)

	// Format JSON content if applicable
	formattedContent := formatContentIfJSON(content, mimeType)

	attachmentPath := filepath.Join(r.targetDir, source)
	if err := os.WriteFile(attachmentPath, []byte(formattedContent), 0644); err != nil {
		return fmt.Errorf("failed to write attachment: %w", err)
	}

	r.Attachments = append(r.Attachments, Attachment{
		Name:   name,
		Source: source,
		Type:   mimeType,
	})

	return nil
}

func (r *Result) Save() error {
	if r.Stop == 0 {
		r.Finish()
	}

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result to JSON: %w", err)
	}

	filename := fmt.Sprintf("%s-result.json", r.UUID)
	resultPath := filepath.Join(r.targetDir, filename)
	if err := os.WriteFile(resultPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write result file: %w", err)
	}

	return nil
}

func currentTimeMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func generateHistoryID(name string) string {
	hash := md5.Sum([]byte(name))
	return hex.EncodeToString(hash[:])
}

func getFileExtension(mimeType string) string {
	switch mimeType {
	case MimeTypeTextPlain:
		return "txt"
	case MimeTypeTextHTML:
		return "html"
	case MimeTypeApplicationJSON:
		return "json"
	case MimeTypeApplicationXML:
		return "xml"
	case MimeTypeImagePNG:
		return "png"
	case MimeTypeImageJPEG:
		return "jpg"
	default:
		return "txt"
	}
}

func NewSeverityLabel(severity string) Label {
	return Label{Name: LabelSeverity, Value: severity}
}

func NewFeatureLabel(feature string) Label {
	return Label{Name: LabelFeature, Value: feature}
}

func NewStoryLabel(story string) Label {
	return Label{Name: LabelStory, Value: story}
}

func NewEpicLabel(epic string) Label {
	return Label{Name: LabelEpic, Value: epic}
}

func NewOwnerLabel(owner string) Label {
	return Label{Name: LabelOwner, Value: owner}
}

func NewTagLabel(tag string) Label {
	return Label{Name: LabelTag, Value: tag}
}

// formatContentIfJSON formats JSON content with indentation if mimeType is application/json.
// If content is not valid JSON, returns the original content unchanged.
func formatContentIfJSON(content, mimeType string) string {
	if mimeType != MimeTypeApplicationJSON {
		return content
	}

	var data interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		// If content is not valid JSON, return as-is
		return content
	}

	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		// If formatting fails, return original content
		return content
	}

	return string(formatted)
}
