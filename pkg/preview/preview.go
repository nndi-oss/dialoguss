package preview

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nndi-oss/dialoguss/pkg/core"
)

func GeneratePreviewAnimation(config *core.Session) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func GenerateScreen(text string, isInput bool) ([]byte, error) {
	return []byte(strings.ReplaceAll(phoneSVGTemplate, "{Placeholder}", text)), nil
}

func GenerateScreens(session *core.Session, outputdir string) (bool, error) {
	for idx, step := range session.Steps {
		rendered, err := GenerateScreen(step.Expect, false)
		if err != nil {
			return false, err
		}
		outfile := filepath.Join(outputdir, fmt.Sprintf("%s_step-%d.svg", session.ID, idx))
		err = os.WriteFile(outfile, rendered, 0o775)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func GenerateGrid() {
	// Generate screens as an image grid, left to right top to bottom
	// TODO: use https://github.com/ozankasikci/go-image-merge
}
