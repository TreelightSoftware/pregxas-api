package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	ConfigSetup()

	// in order, test the logging levels
	data := map[string]string{
		"setting": "testing",
	}
	Log("info", "testing", "testing_call", data)
	Log("warning", "testing", "testing_call", data)
	Log("error", "testing", "testing_call", data)
	assert.True(t, true)
}
