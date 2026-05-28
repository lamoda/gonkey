package models

// AllureMetadata contains metadata for Allure 2 reports and TMS integration
type AllureMetadata struct {
	Links      []AllureLink      `json:"links" yaml:"links"`
	Labels     []AllureLabel     `json:"labels" yaml:"labels"`
	Parameters []AllureParameter `json:"parameters" yaml:"parameters"`
}

type AllureLink struct {
	Name string `json:"name" yaml:"name"`
	URL  string `json:"url" yaml:"url"`
	Type string `json:"type" yaml:"type"`
}

type AllureLabel struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

type AllureParameter struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}
