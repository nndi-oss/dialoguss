package cmd

import "log"

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

	if err := d.Run(); err != nil {
		log.Fatalf("Failed to run dialoguss. Got error %s", err)
	}
	return nil
}
