package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

var (
	interactive    bool
	file           string
	trurouteMode   bool
	defaultTimeout = 21 * time.Second
)

const (
	ApiTypeAfricastalking   = "AT_USSD"
	ApiTypeTruroute         = "TR_USSD"
	InteractiveDialTemplate = `Dialing app using:

	Phone: %s
	Url: %s
	SessionID:%s
	API Type: %s
`
)

/// UnexpectedResultError
///
/// Unexpected result from the USSD application
func UnexpectedResultError(want string, have string) error {
	return fmt.Errorf("Unexpected result.\n\tWant: %s\n\tHave: %s", want, have)
}

type Step struct {
	StepNo int
	isLast bool
	isDial bool
	Text   string `yaml:"userInput"`
	Expect string `yaml:"expect"`
}

/// DialStep
///
/// DialStep is the first step in the session, dials the USSD service
func DialStep(expect string) *Step {
	return &Step{
		StepNo: 0,
		isLast: false,
		isDial: true,
		Text:   "",
		Expect: expect,
	}
}

/// NewStep
///
/// a subsequent step in the session to the USSD service
func NewStep(i int, text string, expect string) *Step {
	return &Step{
		StepNo: i,
		isLast: false,
		isDial: false,
		Text:   text,
		Expect: expect,
	}
}

/// Executes a step and returns the result of the request
/// May return an empty string ("") upon failure
func (s *Step) Execute(session *Session) (string, error) {
	if trurouteMode {
		return s.ExecuteAsTruRouteRequest(session)
	}

	return s.ExecuteAsAfricasTalking(session)
}

type Session struct {
	ID          string  `yaml:"id"`
	PhoneNumber string  `yaml:"phoneNumber"`
	Description string  `yaml:"description"`
	Steps       []*Step `yaml:"steps"`
	serviceCode string
	url         string
	client      *http.Client
	ApiType     string
	Timeout     time.Duration
}

type DialogussConfig struct {
	URL         string    `yaml:"url"`
	Dial        string    `yaml:"dial"`
	PhoneNumber string    `yaml:"phoneNumber"`
	Sessions    []Session `yaml:"sessions"`
	Timeout     int       `yaml:"timeout"`
}

/// AddStep adds step to session
func (s *Session) AddStep(step *Step) {
	s.Steps = append(s.Steps, step)
}

func NewInteractiveSession(d DialogussConfig) *Session {
	rand.Seed(time.Now().UnixNano())
	apiType := ApiTypeAfricastalking
	if trurouteMode {
		apiType = ApiTypeTruroute
	}
	var sessionTimeout = defaultTimeout
	if d.Timeout > 0 {
		sessionTimeout = time.Duration(d.Timeout) * time.Second
	}
	return &Session{
		ID:          fmt.Sprintf("DialogussSession__%d", rand.Uint64()),
		PhoneNumber: d.PhoneNumber,
		Description: "Interactive Session",
		Steps:       nil,
		serviceCode: d.Dial,
		url:         d.URL,
		client:      &http.Client{},
		ApiType:     apiType,
		Timeout:     sessionTimeout,
	}
}

func (s *Session) Run() error {
	first := true
	for i, step := range s.Steps {
		if first {
			step.Execute(s)
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

func prompt() string {
	var s string
	fmt.Print("Enter value> ")
	fmt.Scanln(&s)
	return s
}

// promptCh writes users input into a channel
func promptCh(ch chan string) {
	var value string
	fmt.Print("Enter value> ")
	fmt.Scanln(&value)
	ch <- value
}

func (s *Session) RunInteractive() error {
	var input, output string
	var err error
	var step *Step
	// First Step for the Session is to dial
	step = DialStep("")
	output, err = step.Execute(s)

	apiTypeName := "AfricasTalking USSD"
	if trurouteMode {
		apiTypeName = "TNM TruRoute USSD"
	}

	fmt.Printf(InteractiveDialTemplate,
		s.PhoneNumber,
		s.url,
		s.ID,
		apiTypeName,
	)

	fmt.Println()
	if err != nil {
		return err
	}
	fmt.Println(output)
	// Execute other steps if we haven't received an "END" response
sessionLoop:
	for i := 0; !step.isLast; i++ {
		inputCh := make(chan string, 1)

		// Read the input or timeout after a few seconds (currently 21)
		go promptCh(inputCh)

		select {
		case value := <-inputCh:
			input = value
		case <-time.After(s.Timeout):
			fmt.Println("Session timed out!")
			break sessionLoop
		}

		step = NewStep(i, input, "")
		output, err = step.Execute(s)
		if err != nil {
			return err
		}
		fmt.Println(output)
		if step.isLast {
			break
		}
	}

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
	d.config = DialogussConfig{Timeout: int(defaultTimeout)}
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

	apiType := ApiTypeAfricastalking
	if trurouteMode {
		apiType = ApiTypeTruroute
	}

	for _, session := range d.config.Sessions {
		steps := make([]*Step, len(session.Steps))
		copy(steps, session.Steps)

		s := &Session{
			ID:          session.ID,
			Description: session.Description,
			PhoneNumber: session.PhoneNumber,
			Steps:       steps,
			url:         d.config.URL,
			client:      &http.Client{},
			ApiType:     apiType,
		}

		s.client.Timeout = 10 * time.Second

		go func() {
			defer wg.Done()
			err := s.Run()
			if err != nil {
				//sessionErrors <-fmt.Sprintf("Error in Session %s. Got: %s ", s.ID, err)
				sessionErrors[s.ID] = err
			}
		}()
	}
	wg.Wait()
	for key, val := range sessionErrors {
		log.Printf("Got error in session %s: %s", key, val)
	}
	return nil
}

/// Run executes the sessions
func (d *Dialoguss) Run() error {
	// log.Print("Running dialoguss with config", d.config)
	if d.isInteractive {
		session := NewInteractiveSession(d.config)
		return session.RunInteractive()
	}

	return d.RunAutomatedSessions()
}

func init() {
	flag.BoolVar(&interactive, "i", false, "Interactive")
	flag.BoolVar(&trurouteMode, "truroute-mode", false, "TruRoute USSD mode for developing USSD apps on TNM services")
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
