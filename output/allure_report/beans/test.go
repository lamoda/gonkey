package beans

import (
	"strings"
	"time"
)

// start new test case
func NewTestCase(name string, start time.Time) *TestCase {
	test := new(TestCase)
	test.Name = name

	if !start.IsZero() {
		test.Start = start.UnixNano() / 1000
	} else {
		test.Start = time.Now().UnixNano() / 1000
	}

	return test
}

type TestCase struct {
	Status string `xml:"status,attr"`
	Start  int64  `xml:"start,attr"`
	Stop   int64  `xml:"stop,attr"`
	Name   string `xml:"name"`
	Steps  struct {
		Steps []*Step `xml:"step"`
	} `xml:"steps"`
	Labels struct {
		Label []*Label `xml:"label"`
	} `xml:"labels"`
	Attachments struct {
		Attachment []*Attachment `xml:"attachment"`
	} `xml:"attachments"`
	Desc    string `xml:"description"`
	Failure struct {
		Msg   string `xml:"message"`
		Trace string `xml:"stack-trace"`
	} `xml:"failure,omitempty"`
}

type Label struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

func (t *TestCase) SetDescription(desc string) {
	t.Desc = desc
}

func (t *TestCase) SetDescriptionOrDefaultValue(desc, defVal string) {
	if desc == "" {
		t.Desc = defVal
	} else {
		t.Desc = desc
	}
}

func (t *TestCase) AddLabel(name, value string) {
	t.addLabel(&Label{
		Name:  name,
		Value: value,
	})
}

func (t *TestCase) AddStep(step *Step) {
	t.Steps.Steps = append(t.Steps.Steps, step)
}

func (t *TestCase) AddAttachment(attach *Attachment) {
	t.Attachments.Attachment = append(t.Attachments.Attachment, attach)
}

func (t *TestCase) End(status string, err error, end time.Time) {
	if !end.IsZero() {
		t.Stop = end.UnixNano() / 1000
	} else {
		t.Stop = time.Now().UnixNano() / 1000
	}
	t.Status = status
	if err != nil {
		msg := strings.Split(err.Error(), "\trace")
		t.Failure.Msg = msg[0]
		t.Failure.Trace = strings.Join(msg[1:], "\n")
	}
}

func (t *TestCase) addLabel(label *Label) {
	t.Labels.Label = append(t.Labels.Label, label)
}
