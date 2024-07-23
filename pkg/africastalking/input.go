package africastalking

import (
	"strings"

	"github.com/nndi-oss/dialoguss/pkg/core"
)

func ConcatText(session *core.Session) string {
	if session == nil {
		return ""
	}

	inputs := make([]string, 0)
	for _, step := range session.Steps {
		inputs = append(inputs, step.Text)
	}

	return strings.Join(inputs, "*")
}
