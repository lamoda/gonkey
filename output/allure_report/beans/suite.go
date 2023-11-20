package beans

import (
	"encoding/xml"
	"time"
)

const NsModel = `urn:model.allure.qatools.yandex.ru`

type Suite struct {
	XMLName   xml.Name `xml:"ns2:test-suite"`
	NsAttr    string   `xml:"xmlns:ns2,attr"`
	Start     int64    `xml:"start,attr"`
	End       int64    `xml:"stop,attr"`
	Name      string   `xml:"name"`
	Title     string   `xml:"title"`
	TestCases struct {
		Cases []*TestCase `xml:"test-case"`
	} `xml:"test-cases"`
}

func NewSuite(name string, start time.Time) *Suite {
	s := new(Suite)

	s.NsAttr = NsModel
	s.Name = name
	s.Title = name

	if !start.IsZero() {
		s.Start = start.UTC().UnixNano() / 1000
	} else {
		s.Start = time.Now().UTC().UnixNano() / 1000
	}

	return s
}

// SetEnd set end time for suite
func (s *Suite) SetEnd(endTime time.Time) {
	if !endTime.IsZero() {
		// strict UTC
		s.End = endTime.UTC().UnixNano() / 1000
	} else {
		s.End = time.Now().UTC().UnixNano() / 1000
	}
}

// suite has test-cases?
func (s Suite) HasTests() bool {
	return len(s.TestCases.Cases) > 0
}

// add test in suite
func (s *Suite) AddTest(test *TestCase) {
	s.TestCases.Cases = append(s.TestCases.Cases, test)
}
