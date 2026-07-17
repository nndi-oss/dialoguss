package cmd

import (
	"log"

	"github.com/nndi-oss/dialoguss/pkg/whatsapp"
)

type RunCmd struct {
}

func (r *RunCmd) Run(globals *Globals) error {
	d := &Dialoguss{
		IsInteractive: false,
		File:          globals.File,
	}

	if err := d.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration file. Got error %s", err)
	}

	// If any sessions are configured as WhatsApp, start the mock Meta server
	// User configures their bot's WhatsApp Base API URL to target: http://localhost:8090
	whatsapp.StartMockServer(":8090")

	if err := d.Run(); err != nil {
		log.Fatalf("Failed to run dialoguss. Got error %s", err)
	}
	return nil
}
