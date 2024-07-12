package africastalking_test

import (
	"testing"

	"github.com/nndi-oss/dialoguss/pkg/africastalking"
	"github.com/nndi-oss/dialoguss/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestConcatTextWithNilSession(t *testing.T) {
	got := africastalking.ConcatText(nil)
	assert.Equal(t, "", got)
}

func TestConcatTextWithNilSteps(t *testing.T) {
	got := africastalking.ConcatText(&core.Session{
		Steps: nil,
	})

	assert.Equal(t, "", got)
}

func TestConcatText(t *testing.T) {
	got := africastalking.ConcatText(&core.Session{
		Steps: []*core.Step{
			&core.Step{Text: "1"},
			&core.Step{Text: "2"},
			&core.Step{Text: "3"},
			&core.Step{Text: "4"},
			&core.Step{Text: "5"},
		},
	})

	assert.Equal(t, "1*2*3*4*5", got)
}
