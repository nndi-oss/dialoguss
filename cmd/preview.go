package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/nndi-oss/dialoguss/pkg/core"
)

type PreviewCmd struct {
	ServerURL string `cmd:"" help:"Server URL for a USSD Studio server" default:"https://ussd-studio-api.nndi.cloud"`
}

func (r *PreviewCmd) Run(globals *Globals) error {
	d := &Dialoguss{
		IsInteractive: false,
		File:          globals.File,
	}

	if err := d.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration file. Got error %s", err)
	}

	ussdStudio := &ussdStudioClient{
		baseURL: r.ServerURL,
		Client:  &http.Client{},
	}

	previewURL, err := ussdStudio.GeneratePreview(d.Config)
	if err != nil {
		log.Fatalf("Failed to generate preview from USSD Studio Server. Got error %s", err)
	}

	fmt.Println("Dialoguss USSD Studio Preview Generated")
	fmt.Print("You can open the Preview at this URL\n\n")
	fmt.Printf("\tURL: %s\n", previewURL)

	return nil
}

type ussdStudioClient struct {
	baseURL string
	Client  *http.Client
}

type ussdStudioResponse struct {
	PreviewURL   string `json:"URL"`
	PreviewToken string `json:"Token"`
}

func (studio *ussdStudioClient) GeneratePreview(dialogussConfig core.DialogussConfig) (string, error) {
	apiURL := fmt.Sprintf("%s/api/generate-preview?utm_source=dialoguss-cli&_id=", studio.baseURL)
	data, err := json.Marshal(dialogussConfig)
	if err != nil {
		return "", fmt.Errorf("failed to parse request as JSON:\n%v\nError: %w", string(data), err)
	}

	response, err := studio.Client.Post(apiURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to send request to %s\n Error %w", apiURL, err)
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body from server: %w", err)
	}
	defer response.Body.Close()

	var previewResponse ussdStudioResponse
	err = json.Unmarshal(responseData, &previewResponse)
	if err != nil {
		return "", fmt.Errorf("failed to parse response from server as JSON:\n%v\nError: %w", string(responseData), err)
	}

	// TODO: store the preview token?
	return previewResponse.PreviewURL, nil
}
