package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nndi-oss/dialoguss/pkg/core"
)

func FillComponentStore(store *ComponentStore, namespace string, components ...core.Component) {
	if _, ok := (*store)[namespace]; !ok {
		container := make(map[string]string)
		(*store)[namespace] = container
	}

	for _, component := range components {
		(*store)[namespace][component.ID] = component.Expect
	}
}

func TestResolveStepExpectedValue(t *testing.T) {
	t.Run("Step.Expect is used when no ComponentID is Provided", func(*testing.T) {
		want := "Hello"

		session := core.Session{ID: "foo"}
		step := core.Step{StepNo: 1, Expect: want}
		store := make(ComponentStore)
		FillComponentStore(&store, "default", core.Component{ID: "bar"})

		returned, _ := ResolveStepExpectedValue(&session, &step, &store)
		if returned != want {
			t.Fatalf(`ResolveStepExpectedValue == "%s", want "%s"`, returned, want)
		}
	})

	t.Run("Step.Expect overrides Component.Expect", func(*testing.T) {
		want := "Testicles of Narnia"

		session := core.Session{ID: "foo"}
		step := core.Step{StepNo: 1, Expect: want, ComponentID: "default/bar"}
		store := make(ComponentStore)
		FillComponentStore(&store, "default", core.Component{ID: "bar"})

		returned, _ := ResolveStepExpectedValue(&session, &step, &store)
		if returned != want {
			t.Fatalf(`ResolveStepExpectedValue == "%s", want "%s"`, returned, want)
		}
	})

	t.Run("Component.Expect is used when Step.Expect is not provided", func(t *testing.T) {
		want := "Thou shalt not test the lord thy God"

		session := core.Session{ID: "foo"}
		step := core.Step{StepNo: 1, Expect: "", ComponentID: "default/bar"}
		store := make(ComponentStore)
		FillComponentStore(&store, "default", core.Component{ID: "bar", Expect: want})

		returned, _ := ResolveStepExpectedValue(&session, &step, &store)
		if returned != want {
			t.Fatalf(`ResolveStepExpectedValue == "%s", want "%s"`, returned, want)
		}

	})

	t.Run("Errors when a non existent Step.ComponentID", func(*testing.T) {
		session := core.Session{ID: "foo"}
		step := core.Step{StepNo: 1, Expect: "", ComponentID: "foo/bar"}
		store := make(ComponentStore)
		FillComponentStore(&store, "foo", core.Component{ID: "foo", Expect: "Nada!"})

		_, err := ResolveStepExpectedValue(&session, &step, &store)
		if err == nil {
			t.Fatalf("ResolveStepExpectedValue did not error")
		}
	})

	t.Run("Uses default namespace when Step.ComponentID is not namespaced", func(*testing.T) {
		want := "My patience is being tested here"

		session := core.Session{ID: "foobbar"}
		step := core.Step{StepNo: 1, Expect: "", ComponentID: "bar"}
		store := make(ComponentStore)
		FillComponentStore(&store, "default", core.Component{ID: "bar", Expect: want})

		returned, _ := ResolveStepExpectedValue(&session, &step, &store)
		if returned != want {
			t.Fatalf(`ResolveStepExpectedValue == "%s", want "%s"`, returned, want)
		}
	})
}

func CreateTestDialoguss(url string, steps []*core.Step, components []*core.Component) *Dialoguss {
	return &Dialoguss{
		IsInteractive: false,
		File:          "/dev/null",
		Config: core.DialogussConfig{
			URL:         url,
			Dial:        "*6969#",
			PhoneNumber: "0888800900",
			Sessions: []core.Session{
				{
					ID:          "c-session",
					PhoneNumber: "0888800900",
					Description: "It's the gat damn session",
					Steps:       steps,
					Url:         url,
					Client:      &http.Client{},
					ApiType:     ApiTypeAfricastalking,
					Timeout:     1000,
				},
			},
			Components: []core.ComponentNamespace{
				{
					Namespace: "default",
					Items:     components,
				},
			},
		},
	}
}

func CreateMockHttpServer(status int, responses ...string) *httptest.Server {
	responseNo := 0
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if responseNo >= len(responses) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("You are asking too many questions!"))
			return
		}

		anchor := "CON"
		if responseNo < len(responses)-1 {
			anchor = "END"
		}

		response := fmt.Sprintf("%s %s", anchor, responses[responseNo])

		w.WriteHeader(status)
		w.Write([]byte(response))
		responseNo += 1
	}))
}

func TestRunAutomatedSession(t *testing.T) {
	t.Run("Passes steps with matched expectations", func(*testing.T) {
		server := CreateMockHttpServer(http.StatusOK, "Welcome", "Hello")
		defer server.Close()

		dialoguss := CreateTestDialoguss(
			server.URL,
			[]*core.Step{
				{
					Expect: "Welcome",
				},
				{
					Text:   "1",
					Expect: "Hello",
				},
			},
			[]*core.Component{},
		)
		failed, err := dialoguss.RunAutomatedSessions()
		if err != nil {
			t.Fatalf("Session run failed: %v", err)
		}

		if failed != 0 {
			t.Fatalf("Unexpected number of failed sessions, expected: 0, got: %d", failed)
		}
	})

	t.Run("Fails steps with unmatched expectations", func(*testing.T) {
		server := CreateMockHttpServer(http.StatusOK, "Welcome", "Hello")
		defer server.Close()

		dialoguss := CreateTestDialoguss(
			server.URL,
			[]*core.Step{
				{
					Expect: "Mwalandilidwa",
				},
				{
					Expect: "Hello",
					Text:   "2",
				},
			},
			[]*core.Component{},
		)
		failed, err := dialoguss.RunAutomatedSessions()
		if err != nil {
			t.Fatalf("Session run failed: %v", err)
		}

		if failed != 1 {
			t.Fatalf("Unexpected number of failed sessions, expected: 1, got: %d", failed)
		}
	})

	t.Run("Passes steps with matching component expectations", func(*testing.T) {
		server := CreateMockHttpServer(http.StatusOK, "Welcome", "Hello")
		defer server.Close()

		dialoguss := CreateTestDialoguss(
			server.URL,
			[]*core.Step{
				{
					ComponentID: "default.splash",
				},
				{
					ComponentID: "home",
					Expect:      "Hello",
				},
			},
			[]*core.Component{
				{
					ID:     "splash",
					Expect: "Welcome",
				},
				{
					ID:     "home",
					Expect: "Hello",
				},
			},
		)
		failed, err := dialoguss.RunAutomatedSessions()
		if err != nil {
			t.Fatalf("Session run failed: %v", err)
		}

		if failed != 0 {
			t.Fatalf("Unexpected number of failed sessions, expected: 0, got: %d", failed)
		}
	})

	t.Run("Fails steps with non-matching component expectations", func(t *testing.T) {
		server := CreateMockHttpServer(http.StatusOK, "Welcome", "Hello")
		defer server.Close()

		dialoguss := CreateTestDialoguss(
			server.URL,
			[]*core.Step{
				{
					ComponentID: "default.splash",
				},
				{
					ComponentID: "home",
					Expect:      "Hello",
				},
			},
			[]*core.Component{
				{
					ID:     "splash",
					Expect: "Mwalandilidwa",
				},
				{
					ID:     "home",
					Expect: "Hello",
				},
			},
		)
		failed, err := dialoguss.RunAutomatedSessions()
		if err != nil {
			t.Fatalf("Session run failed: %v", err)
		}

		if failed != 1 {
			t.Fatalf("Unexpected number of failed sessions, expected: 1, got: %d", failed)
		}
	})
}
