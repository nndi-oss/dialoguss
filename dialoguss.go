package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

var (
	interactive bool
	file        string
	reUssdCon   = regexp.MustCompile(`^CON\s?`)
	reUssdEnd   = regexp.MustCompile(`^END\s?`)
)

/// UnexpectedResultError
///
/// Unexpected result from the USSD application
func UnexpectedResultError(want string, have string) error {
	return fmt.Errorf("Unexpected result.\n\tWant: %s\n\tHave: %s", want, have)
}

type Step struct {
	StepNo  int
	isLast  bool
	isDial  bool
	Text    string `yaml:"text"`
	Expect  string `yaml:"expect"`
	session *Session
}

/// DialStep
///
/// DialStep is the first step in the session, dials the USSD service
func DialStep(expect string) Step {
	return Step{
		StepNo:  0,
		isLast:  false,
		isDial:  true,
		Text:    "",
		Expect:  expect,
		session: nil,
	}
}

/// NewStep
///
/// a subsequent step in the session to the USSD service
func NewStep(i int, text string, expect string) Step {
	return Step{
		StepNo:  i,
		isLast:  false,
		isDial:  false,
		Text:    text,
		Expect:  expect,
		session: nil,
	}
}

/// Executes a step and returns the result of the request
/// May return an empty string ("") upon failure
func (s *Step) Execute(session *Session) (string, error) {
	data := url.Values{}
	data.Set("sessionId", session.ID)
	data.Set("phoneNumber", session.PhoneNumber)
	data.Set("text", s.Text) // TODO(zikani): concat the input
	data.Set("channel", "")  // TODO: Get the channel

	res, err := session.client.PostForm(session.url, data)
	if err != nil {
		log.Printf("Failed to send request to %s", session.url)
		return "", err
	}

	b, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return "", err
	}

	responseText := string(b)
	if reUssdCon.MatchString(responseText) {
		responseText = strings.Replace(responseText, "CON ", "", 1)
		s.isLast = false
	} else if reUssdEnd.MatchString(responseText) {
		responseText = strings.Replace(responseText, "END ", "", 1)
		s.isLast = true
	}

	return responseText, nil
}

type Session struct {
	ID          string `yaml:"id"`
	PhoneNumber string `yaml:"phoneNumber"`
	Description string `yaml:"description"`
	Steps       []Step `yaml:"steps"`
	url         string
	client      *http.Client
}

type DialogussConfig struct {
	URL         string    `yaml:"url"`
	Dial        string    `yaml:"dial"`
	PhoneNumber string    `yaml:"phoneNumber"`
	Sessions    []Session `yaml:"sessions"`
}

/// AddStep adds step to session
func (s *Session) AddStep(step Step) {
	s.Steps = append(s.Steps, step)
}

func NewInteractiveSession(d DialogussConfig) *Session {
	return &Session{
		ID:          string(rand.Intn(99999)),
		PhoneNumber: d.PhoneNumber,
		Description: "Interactive Session",
		Steps:       nil,
		url:         d.URL,
		client:      &http.Client{},
	}
}

func (s *Session) Run() error {
	first := true
	for i, step := range s.Steps {
		if first {
			DialStep(step.Expect).Execute(s)
			first = false
			continue
		}
		step.StepNo = i
		result, err := step.Execute(s)
		if err != nil {
			log.Printf("Failed to execute step %d", step.StepNo)
			return err
		}
		if result != step.Expect {
			return UnexpectedResultError(step.Expect, result)
		}
	}
	log.Printf("All steps in session %s run successfully", s.ID)
	return nil
}

/// Dialoguss
///
/// Dialoguss is an application that can have one or more pseudo-ussd sessions
type Dialoguss struct {
	isInteractive bool
	file          string
	config        DialogussConfig
}

/// LoadConfig loads configuration from YAML
func (d *Dialoguss) LoadConfig() error {
	d.config = DialogussConfig{}
	b, err := ioutil.ReadFile(d.file)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, &d.config)
}

/// Loads the sessions for this application
func (d *Dialoguss) RunAutomatedSessions() error {
	var wg sync.WaitGroup
	wg.Add(len(d.config.Sessions))
	sessionErrors := make(map[string]error)

	for _, session := range d.config.Sessions {
		s := &session
		//go func() {
		s.client = &http.Client{}
		s.client.Timeout = 10 * time.Second
		s.url = d.config.URL
		err := s.Run()
		if err != nil {
			sessionErrors[s.ID] = err
		}
		wg.Done()
		//}()
	}
	wg.Wait()
	// TODO: collect errors and return any here
	for key, val := range sessionErrors {
		log.Printf("SessionID=%s Got error=%s", key, val)
	}
	return nil
}

/// Run executes the sessions
func (d *Dialoguss) Run() error {
	// log.Print("Running dialoguss with config", d.config)
	if d.isInteractive {
		session := NewInteractiveSession(d.config)
		return session.Run()
	}

	return d.RunAutomatedSessions()
}

func init() {
	flag.BoolVar(&interactive, "i", false, "Interactive")
	flag.StringVar(&file, "f", "dialoguss.yml", "Dialoguss configuration file")
}

func main() {
	flag.Parse()
	d := &Dialoguss{
		isInteractive: interactive,
		file:          file,
	}

	if err := d.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration file. Got error %s", err)
	}

	if err := d.Run(); err != nil {
		log.Fatalf("Failed to run dialoguss. Got error %s", err)
	}
}
