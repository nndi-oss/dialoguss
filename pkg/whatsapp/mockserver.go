package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/nndi-oss/dialoguss/pkg/core"
)

// MockMetaServer captures outgoing WhatsApp API calls from the chatbot.
type MockMetaServer struct {
	server *http.Server
	mu     sync.Mutex
	// Maps Recipient Phone -> Channel of received message strings
	channels map[string]chan string
}

var MockServer *MockMetaServer

func StartMockServer(addr string) {
	MockServer = &MockMetaServer{
		channels: make(map[string]chan string),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v20.0/", MockServer.handleMessageSend)
	MockServer.server = &http.Server{Addr: addr, Handler: mux}
	go MockServer.server.ListenAndServe()
}

func (m *MockMetaServer) handleMessageSend(w http.ResponseWriter, r *http.Request) {
	// Parse the outgoing WhatsApp API call (the bot's reply)
	// Expects standard WhatsApp JSON payloads (like the Client we created earlier)
	var payload struct {
		To   string `json:"to"`
		Text struct {
			Body string `json:"body"`
		} `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	m.mu.Lock()
	ch, exists := m.channels[payload.To]
	m.mu.Unlock()

	if exists {
		ch <- payload.Text.Body
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"messaging_product": "whatsapp", "messages": [{"id": "wamid.mock"}]}`))
}

func (m *MockMetaServer) ExpectReply(phone string, timeout time.Duration) (string, error) {
	m.mu.Lock()
	ch, exists := m.channels[phone]
	if !exists {
		ch = make(chan string, 1)
		m.channels[phone] = ch
	}
	m.mu.Unlock()

	select {
	case reply := <-ch:
		return reply, nil
	case <-time.After(timeout):
		return "", fmt.Errorf("timeout waiting for chatbot reply")
	}
}

type WhatsAppDriver struct {
	Step *core.Step
}

func (w WhatsAppDriver) Execute(session *core.Session) (string, error) {
	// 1. Build Meta webhook payload representing a message from the user
	payload := map[string]interface{}{
		"object": "whatsapp_business_account",
		"entry": []interface{}{
			map[string]interface{}{
				"id": "dialoguss__" + string(time.Now().UnixMilli()),
				"changes": []interface{}{
					map[string]interface{}{
						"value": map[string]interface{}{
							"messaging_product": "whatsapp",
							"contacts": []interface{}{
								map[string]interface{}{
									"wa_id":   session.PhoneNumber,
									"profile": map[string]interface{}{"name": "Test User"},
								},
							},
							"messages": []interface{}{
								map[string]interface{}{
									"from":      session.PhoneNumber,
									"id":        "wamid.testmessageid",
									"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
									"text":      map[string]interface{}{"body": w.Step.Text},
									"type":      "text",
								},
							},
						},
						"field": "messages",
					},
				},
			},
		},
	}

	body, _ := json.Marshal(payload)

	// 2. Register expectation with Mock Server before sending Webhook
	replyChan := make(chan string, 1)
	MockServer.mu.Lock()
	MockServer.channels[session.PhoneNumber] = replyChan
	MockServer.mu.Unlock()

	// 3. Send Webhook to chatbot
	req, _ := http.NewRequest("POST", session.Url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := session.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to deliver webhook: %w", err)
	}
	resp.Body.Close()

	// 4. Wait for the chatbot to make an API request back to our MockServer
	return MockServer.ExpectReply(session.PhoneNumber, session.Timeout)
}
