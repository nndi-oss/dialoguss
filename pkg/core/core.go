package core

import (
	"net/http"
	"time"
)

// Dialoguss is an application that can have one or more pseudo-ussd sessions
type Dialoguss struct {
	IsInteractive bool
	File          string
	Config        DialogussConfig
}

type Step struct {
	StepNo      int
	IsLast      bool
	IsDial      bool
	Text        string `yaml:"userInput"`
	Expect      string `yaml:"expect"`
	ComponentID string `yaml:"componentId"`
}

type Component struct {
	ID     string `yaml:"id"`
	Expect string `yaml:"expect"`
}

type ComponentNamespace struct {
	Namespace string       `yaml:"namespace"`
	Items     []*Component `yaml:"component"`
}

type Session struct {
	ID          string  `yaml:"id"`
	PhoneNumber string  `yaml:"phoneNumber"`
	Description string  `yaml:"description"`
	Steps       []*Step `yaml:"steps"`
	ServiceCode string
	Url         string
	Client      *http.Client
	ApiType     string
	Timeout     time.Duration
}

type DialogussConfig struct {
	URL         string               `yaml:"url"`
	Dial        string               `yaml:"dial"`
	PhoneNumber string               `yaml:"phoneNumber"`
	Sessions    []Session            `yaml:"sessions"`
	Components  []ComponentNamespace `yaml:"components"`
	Timeout     int                  `yaml:"timeout"`
}
