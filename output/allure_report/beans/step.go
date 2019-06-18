package beans

import (
	"time"
)

func NewStep(name string, start time.Time) *Step {
	test := new(Step)
	test.Name = name

	if !start.IsZero() {
		test.Start = start.UnixNano() / 1000
	} else {
		test.Start = time.Now().UnixNano() / 1000
	}

	return test
}

type Step struct {
	Parent *Step `xml:"-"`

	Status      string        `xml:"status,attr"`
	Start       int64         `xml:"start,attr"`
	Stop        int64         `xml:"stop,attr"`
	Name        string        `xml:"name"`
	Steps       []*Step       `xml:"steps"`
	Attachments []*Attachment `xml:"attachments"`
}

func (s *Step) End(status string, end time.Time) {
	if !end.IsZero() {
		s.Stop = end.UnixNano() / 1000
	} else {
		s.Stop = time.Now().UnixNano() / 1000
	}
	s.Status = status
}

func (s *Step) AddStep(step *Step) {
	if step != nil {
		s.Steps = append(s.Steps, step)
	}
}
