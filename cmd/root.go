package cmd

import (
	"github.com/alecthomas/kong"
)

var version string = "0.7.0-alpha"

type VersionFlag string

type Globals struct {
	File        string `help:"Location of configuration file" default:"dialoguss.yaml" short:"f" type:"path"`
	Interactive bool   `help:"Interactive mode (similar to running dialoguss run -f ...)"`
	Debug       bool   `help:"Enable debug mode"`
	// Version     VersionFlag `name:"version" help:"Show version and quit"`
}

type DialogussCLI struct {
	Globals

	Preview  PreviewCmd  `cmd:"" help:"Generate a preview of the dialoguss sessions using USSD Studio"`
	Run      RunCmd      `cmd:"" help:"Run a dialoguss test from a file"`
	Simulate SimulateCmd `cmd:"" help:"Simulate interaction with a USSD server in the command-line from a file"`
}

func Execute() {
	cli := DialogussCLI{
		Globals: Globals{
			// Version: VersionFlag(version),
		},
	}

	ctx := kong.Parse(&cli,
		kong.Name("dialoguss"),
		kong.Description("`dialoguss` is a cli tool to test USSD applications that are implemented as HTTP services"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Vars{
			"version": version,
		})

	err := ctx.Run(&cli.Globals)
	ctx.FatalIfErrorf(err)
}
