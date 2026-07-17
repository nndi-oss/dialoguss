package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://graph.facebook.com/v20.0"
)

// Client represents the WhatsApp Business Cloud API client.
type Client struct {
	baseURL      string
	phoneNumberID string
	accessToken  string
	httpClient   *http.Client
}

// NewClient creates a new WhatsApp API client.
func NewClient(phoneNumberID, accessToken string) *Client {
	return &Client{
		baseURL:      defaultBaseURL,
		phoneNumberID: phoneNumberID,
		accessToken:  accessToken,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SetBaseURL overrides the default Graph API base URL (useful for testing/mocking).
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

// MessageResponse represents the response payload from WhatsApp API.
type MessageResponse struct {
	MessagingProduct string `json:"messaging_product"`
	Contacts         []struct {
		Input string `json:"input"`
		WaID  string `json:"wa_id"`
	} `json:"contacts"`
	Messages []struct {
		ID string `json:"id"`
	} `json:"messages"`
}

// ErrorResponse represents an API error response.
type ErrorResponse struct {
	Error struct {
		Message   string `json:"message"`
		Type      string `json:"type"`
		Code      int    `json:"code"`
		FBTraceID string `json:"fbtrace_id"`
	} `json:"error"`
}

type textPayload struct {
	MessagingProduct string      `json:"messaging_product"`
	RecipientType    string      `json:"recipient_type"`
	To               string      `json:"to"`
	Type             string      `json:"type"`
	Text             textBody    `json:"text"`
}

type textBody struct {
	PreviewURL bool   `json:"preview_url"`
	Body       string `json:"body"`
}

type templatePayload struct {
	MessagingProduct string           `json:"messaging_product"`
	RecipientType    string           `json:"recipient_type"`
	To               string           `json:"to"`
	Type             string           `json:"type"`
	Template         templateDetails  `json:"template"`
}

type templateDetails struct {
	Name     string           `json:"name"`
	Language templateLanguage `json:"language"`
}

type templateLanguage struct {
	Code string `json:"code"`
}

// SendTextMessage sends a free-form text message to a specific recipient.
// Note: This only works if the 24-hour customer service window is active.
func (c *Client) SendTextMessage(to string, text string) (*MessageResponse, error) {
	payload := textPayload{
		MessagingProduct: "whatsapp",
		RecipientType:    "individual",
		To:               to,
		Type:             "text",
		Text: textBody{
			PreviewURL: false,
			Body:       text,
		},
	}
	return c.send(payload)
}

// SendTemplateMessage sends a pre-approved template message to a specific recipient.
// Required to initiate a message thread if no active session exists.
func (c *Client) SendTemplateMessage(to string, templateName string, languageCode string) (*MessageResponse, error) {
	payload := templatePayload{
		MessagingProduct: "whatsapp",
		RecipientType:    "individual",
		To:               to,
		Type:             "template",
		Template: templateDetails{
			Name: templateName,
			Language: templateLanguage{
				Code: languageCode,
			},
		},
	}
	return c.send(payload)
}

func (c *Client) send(payload interface{}) (*MessageResponse, error) {
	url := fmt.Sprintf("%s/%s/messages", c.baseURL, c.phoneNumberID)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("whatsapp api error (code %d): %s [trace: %s]",
				errResp.Error.Code, errResp.Error.Message, errResp.Error.FBTraceID)
		}
		return nil, fmt.Errorf("whatsapp api responded with status %d: %s", resp.StatusCode, string(respBody))
	}

	var msgResp MessageResponse
	if err := json.Unmarshal(respBody, &msgResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &msgResp, nil
}
