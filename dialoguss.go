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

	"github.com/nndi-oss/dialoguss/pkg/africastalking"
	"github.com/nndi-oss/dialoguss/pkg/core"
	"github.com/nndi-oss/dialoguss/pkg/trueroute"
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

// Dialoguss the main type for interacting with dialoguss sessions
type Dialoguss core.Dialoguss

// UnexpectedResultError unexpected result from the USSD application
func UnexpectedResultError(want string, have string) error {
	return fmt.Errorf("Unexpected result.\n\tWant: %s\n\tHave: %s", want, have)
}

// DialStep is the first step in the session, dials the USSD service
func DialStep(expect string) *core.Step {
	return &core.Step{
		StepNo: 0,
		IsLast: false,
		IsDial: true,
		Text:   "",
		Expect: expect,
	}
}

// NewStep a subsequent step in the session to the USSD service
func NewStep(i int, text string, expect string) *core.Step {
	return &core.Step{
		StepNo: i,
		IsLast: false,
		IsDial: false,
		Text:   text,
		Expect: expect,
	}
}

// Execute executes a step and returns the result of the request may return an empty string ("") upon failure
func Execute(s *core.Step, session *core.Session) (string, error) {
	if trurouteMode {
		step := trueroute.TrueRouteStep{Step: s}
		return step.ExecuteAsTruRouteRequest(session)
	}

	step := africastalking.AfricasTalkingRouteStep{Step: s}
	return step.ExecuteAsAfricasTalking(session)
}

// AddStep adds step to session
func AddStep(s *core.Session, step *core.Step) {
	s.Steps = append(s.Steps, step)
}

// NewInteractiveSession creates a new interactive session that should be run using RunInteractive
func NewInteractiveSession(d core.DialogussConfig) *core.Session {
	rand.Seed(time.Now().UnixNano())
	apiType := ApiTypeAfricastalking
	if trurouteMode {
		apiType = ApiTypeTruroute
	}
	sessionTimeout := defaultTimeout
	if d.Timeout > 0 {
		sessionTimeout = time.Duration(d.Timeout) * time.Second
	}
	return &core.Session{
		ID:          fmt.Sprintf("DialogussSession__%d", rand.Uint64()),
		PhoneNumber: d.PhoneNumber,
		Description: "Interactive Session",
		Steps:       nil,
		ServiceCode: d.Dial,
		Url:         d.URL,
		Client:      &http.Client{},
		ApiType:     apiType,
		Timeout:     sessionTimeout,
	}
}

// Run runs the dialoguss session and executes the steps in each session
func Run(s *core.Session) error {
	first := true
	for i, step := range s.Steps {
		if first {
			Execute(step, s)
			first = false
			continue
		}
		step.StepNo = i
		result, err := Execute(step, s)
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

// RunInteractive runs and interactive session that prompts ufor user input
func RunInteractive(s *core.Session) error {
	var input, output string
	var err error
	var step *core.Step
	// First Step for the Session is to dial
	step = DialStep("")
	output, err = Execute(step, s)

	apiTypeName := "AfricasTalking USSD"
	if trurouteMode {
		apiTypeName = "TNM TruRoute USSD"
	}

	fmt.Printf(InteractiveDialTemplate,
		s.PhoneNumber,
		s.Url,
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
	for i := 0; !step.IsLast; i++ {
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
		output, err = Execute(step, s)
		if err != nil {
			return err
		}
		fmt.Println(output)
		if step.IsLast {
			break
		}
	}

	return nil
}

// LoadConfig loads configuration from YAML
func (d *Dialoguss) LoadConfig() error {
	d.Config = core.DialogussConfig{Timeout: int(defaultTimeout)}
	b, err := ioutil.ReadFile(d.File)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, &d.Config)
}

// RunAutomatedSessions Loads the sessions for this application
func (d *Dialoguss) RunAutomatedSessions() error {
	var wg sync.WaitGroup
	wg.Add(len(d.Config.Sessions))

	sessionErrors := make(map[string]error)

	apiType := ApiTypeAfricastalking
	if trurouteMode {
		apiType = ApiTypeTruroute
	}

	for _, session := range d.Config.Sessions {
		steps := make([]*core.Step, len(session.Steps))
		copy(steps, session.Steps)

		s := &core.Session{
			ID:          session.ID,
			Description: session.Description,
			PhoneNumber: session.PhoneNumber,
			Steps:       steps,
			Url:         d.Config.URL,
			Client:      &http.Client{},
			ApiType:     apiType,
		}

		s.Client.Timeout = 10 * time.Second

		go func() {
			defer wg.Done()
			err := Run(s)
			if err != nil {
				// sessionErrors <-fmt.Sprintf("Error in Session %s. Got: %s ", s.ID, err)
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

// Run executes the sessions
func (d *Dialoguss) Run() error {
	// log.Print("Running dialoguss with config", d.config)
	if d.IsInteractive {
		session := NewInteractiveSession(d.Config)
		return RunInteractive(session)
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
		IsInteractive: interactive,
		File:          file,
	}

	if err := d.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration file. Got error %s", err)
	}

	if err := d.Run(); err != nil {
		log.Fatalf("Failed to run dialoguss. Got error %s", err)
	}
}
