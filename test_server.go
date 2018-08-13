package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"bitbucket.org/nndi/phada"
)

const (
	DEFAULT_CHANNEL      = "384"
	STATE_NOOP           = -1
	STATE_PROMPT_NAME    = 0
	STATE_MENU           = 1
	STATE_ACCOUNT_DETAIL = 2
	STATE_BALANCE        = 3
	STATE_SOMETHING_ELSE = 4
)

type ussdHandler struct {
	http.Handler
	sessionStore phada.SessionStore
}

func (u *ussdHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	session, err := phada.ParseUssdRequest(req)
	if err != nil {
		log.Printf("Failed to parse UssdRequest from http.Request. Error %s", err)
		fmt.Fprintf(w, ussdEnd("Failed to process request"))
		return
	}
	log.Printf("Got the following request text: %s", session.ReadIn())
	session.SetState(STATE_NOOP)
	u.sessionStore.PutHop(session)
	// read the session from the store
	session, err = u.sessionStore.Get(session.SessionID)
	// log.Printf("Have the following session text: %s", session.ReadIn())
	if err != nil {
		log.Printf("Failed to read session %s", err)
		fmt.Fprintf(w, ussdEnd("Failed to process request"))
		return
	}

	if session.ReadIn() == "" || session.State == STATE_NOOP {
		session.SetState(STATE_PROMPT_NAME)
	}

	if session.State == STATE_PROMPT_NAME {
		session.SetState(STATE_MENU)
	}

	if session.State == STATE_MENU && session.ReadIn() == "1" {
		session.SetState(STATE_ACCOUNT_DETAIL)
	}

	if session.State == STATE_MENU && session.ReadIn() == "2" {
		session.SetState(STATE_BALANCE)
	}

	if session.State == STATE_MENU && session.ReadIn() == "3" {
		session.SetState(STATE_SOMETHING_ELSE)
	}

	switch session.State {
	case STATE_PROMPT_NAME:
		fmt.Fprintf(w, ussdContinue("What is your name?"))
		break
	case STATE_MENU:
		s := `Welcome, %s
Choose an item:
1. Account detail
2. Balance
3. Something else
# Exit
`
		fmt.Fprintf(w, ussdContinue(fmt.Sprintf(s, session.Text)))
		break
	case STATE_ACCOUNT_DETAIL:
		fmt.Fprintf(w, ussdEnd("Your account is inactive"))
		break
	case STATE_BALANCE:
		fmt.Fprintf(w, ussdEnd("Your balance is: MK 500"))
		break
	case STATE_SOMETHING_ELSE:
		fmt.Fprintf(w, ussdEnd("There is nothing else here :D"))
		break
	case STATE_NOOP:
	default:
		fmt.Fprintf(w, ussdEnd("Bye"))
		break
	}
}

/// CreateHTTPServer
///
/// DO NOT use this as an example of what an actual USSD
// request handling server should be implemented, please.
func CreateHTTPServer() *http.Server {
	handler := &ussdHandler{
		sessionStore: phada.NewInMemorySessionStore(),
	}
	server := &http.Server{
		Addr:        ":7654",
		Handler:     handler,
		ReadTimeout: 10 * time.Second,
	}

	return server
}

func ussdContinue(text string) string {
	return fmt.Sprintf("CON %s", text)
}

func ussdEnd(text string) string {
	return fmt.Sprintf("END %s", text)
}

func main() {
	s := CreateHTTPServer()
	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start ussd server")
	}
}
