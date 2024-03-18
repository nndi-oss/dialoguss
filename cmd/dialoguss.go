package cmd

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
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

type ComponentStore map[string]map[string]string // namespace => component_id => expect

// UnexpectedResultError unexpected result from the USSD application
func UnexpectedResultError(want string, have string) error {
	return fmt.Errorf("Unexpected result.\n\tWant: %s\n\tHave: %s", want, have)
}

func InvalidComponentIdError(step *core.Step, session *core.Session) error {
	return fmt.Errorf("Invalid component id on step %d of session %s", step.StepNo, session.ID)
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
func ExecuteStep(s *core.Step, session *core.Session) (string, error) {
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

// ResolveStepExpectedValue extracts step's expected value from the step or component store
func ResolveStepExpectedValue(s *core.Session, step *core.Step, store *ComponentStore) (string, error) {
	expect := step.Expect

	if len(expect) > 0 || len(step.ComponentID) == 0 {
		return expect, nil
	}

	path := strings.SplitN(step.ComponentID, "/", 2)
	namespace := "default"
	component_id := ""

	if len(path) == 2 {
		namespace = path[0]
		component_id = path[1]
	} else {
		component_id = path[0]
	}

	container, ok := (*store)[namespace]
	if !ok {
		return "", InvalidComponentIdError(step, s)
	}

	expect, ok = container[component_id]
	if !ok {
		return "", InvalidComponentIdError(step, s)
	}

	return expect, nil
}

// Run runs the dialoguss session and executes the steps in each session
func Run(s *core.Session, store *ComponentStore) error {
	for i, step := range s.Steps {
		step.StepNo = i
		result, err := ExecuteStep(step, s)
		if err != nil {
			log.Printf("Failed to execute step %d", step.StepNo)
			return err
		}

		expect, err := ResolveStepExpectedValue(s, step, store)
		if err != nil {
			return err
		}

		if result != expect {
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
	output, err = ExecuteStep(step, s)

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
		output, err = ExecuteStep(step, s)
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
	b, err := os.ReadFile(d.File)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, &d.Config)
}

func (d *Dialoguss) GetComponents() (*ComponentStore, error) {
	store := make(ComponentStore)
	if d.Config.Components == nil {
		return &store, nil
	}

	id_pattern := regexp.MustCompile(`^[A-Za-z]+[A-Za-z0-9_-]*$`)

	for _, container := range d.Config.Components {
		if !id_pattern.MatchString(container.Namespace) {
			return nil, fmt.Errorf("Invalid component namespace: %s, expected alphanumeric chars only", container.Namespace)
		}

		components, ok := store[container.Namespace]
		if !ok {
			components = make(map[string]string)
		}

		for _, component := range container.Items {
			if !id_pattern.MatchString(component.ID) {
				return nil, fmt.Errorf("Invalid component ID: %s, expected alphanumeric chars only", component.ID)
			}

			components[component.ID] = component.Expect
		}

		store[container.Namespace] = components
	}

	return &store, nil
}

// RunAutomatedSessions Loads the sessions for this application
//
// Returns number of failed sessions
func (d *Dialoguss) RunAutomatedSessions() (int, error) {
	var wg sync.WaitGroup
	wg.Add(len(d.Config.Sessions))

	sessionErrors := make(map[string]error)

	apiType := ApiTypeAfricastalking
	if trurouteMode {
		apiType = ApiTypeTruroute
	}

	components, err := d.GetComponents()
	if err != nil {
		return 0, err
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
			err := Run(s, components)
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
	return len(sessionErrors), nil
}

// Run executes the sessions
func (d *Dialoguss) Run() error {
	// log.Print("Running dialoguss with config", d.config)
	if d.IsInteractive {
		session := NewInteractiveSession(d.Config)
		return RunInteractive(session)
	}

	_, err := d.RunAutomatedSessions()
	return err
}
