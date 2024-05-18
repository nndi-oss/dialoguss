package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/nndi-oss/dialoguss/pkg/core"
	"github.com/nndi-oss/dialoguss/pkg/preview"
)

type PreviewCmd struct {
	OutputDir string `cmd:"" default:"dialoguss_preview" help:"Output directory, defaults 'dialoguss_preview'"`
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
		outputDir: r.OutputDir,
	}

	folder, err := ussdStudio.GeneratePreview(d.Config)
	if err != nil {
		log.Fatalf("Failed to generate preview from USSD Studio Server. Got error %s", err)
	}

	fmt.Println("Dialoguss USSD Studio Preview Generated")
	fmt.Print("You can open the Preview at this URL\n\n")
	fmt.Printf("\tDirectory: %s\n", folder)

	return nil
}

type ussdStudioClient struct {
	outputDir string
}

func (studio *ussdStudioClient) GeneratePreview(dialogussConfig core.DialogussConfig) (string, error) {
	outdir := "./dialoguss_preview"
	if studio.outputDir != "" {
		if _, err := os.Stat(studio.outputDir); os.IsNotExist(err) {
			mkErr := os.Mkdir(studio.outputDir, 0o755)
			if mkErr != nil {
				return "", fmt.Errorf("directory does not exist and could not be created %w", mkErr)
			}
		}
		outdir = studio.outputDir
	}

	for _, session := range dialogussConfig.Sessions {
		sessionOutdir := filepath.Join(outdir, session.ID)
		err := os.Mkdir(sessionOutdir, 0o775)
		if err != nil {
			return "", nil
		}
		_, err = preview.GenerateScreens(&session, sessionOutdir)
		if err != nil {
			return "", nil
		}
	}
	return outdir, nil
}
