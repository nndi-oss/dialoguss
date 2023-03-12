package main

import (
	"testing"

	"github.com/nndi-oss/dialoguss/cmd"

	"github.com/stretchr/testify/assert"
)

func TestDialStep(t *testing.T) {
	dialStep := cmd.DialStep("expected")

	assert.NotNil(t, dialStep, "DialStep should return non nil")
	assert.Equal(t, "", dialStep.Text)
	assert.Equal(t, "expected", dialStep.Expect)
}

func TestNewStep(t *testing.T) {
	newStep := cmd.NewStep(0, "input", "expected")

	assert.Equal(t, 0, newStep.StepNo)
	assert.Equal(t, "input", newStep.Text)
	assert.Equal(t, "expected", newStep.Expect)
	assert.False(t, newStep.IsDial)
	assert.False(t, newStep.IsLast)
}
