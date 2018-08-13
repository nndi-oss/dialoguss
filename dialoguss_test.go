package main

import (
	"log"
	"net/http"
	"testing"
)

func TestShouldRunAutomatedSession(t *testing.T) {
	ch := make(chan bool, 1)
	go func() {
		server := CreateHTTPServer()
		err := server.ListenAndServe()
		if err != nil {
			log.Fatalf("Failed to run the http server. Error %s", err)
		}
		select {
		case <-ch:
			server.Close()
		default:
		}
	}()

	steps := make([]*Step, 3)
	steps = append(steps, DialStep("What is your name?"))
	steps = append(steps, NewStep(1, "Zikani", `Welcome, Zikani
Choose an item:
1. Account detail
2. Balance
3. Something else
# Exit
`,
	))
	steps = append(steps, NewStep(2, "2", "Your balance is: MK 500"))

	auto := &Session{
		ID:          "testSession",
		PhoneNumber: "265888123456",
		Description: "Test Session for dialoguss",
		url:         "http://localhost:7654",
		Steps:       steps,
		client:      &http.Client{},
	}
	// TODO: add test assertions here
	auto.Run()
	ch <- true
}
