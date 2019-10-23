package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendEmail(t *testing.T) {
	ConfigSetup()
	resp, id, err := SendEmail("", "", "")
	// in a testing environment, we expect these to be blank, and a nil error
	assert.Equal(t, "", resp)
	assert.Equal(t, "", id)
	assert.Nil(t, err)
	resp, id, err = SendEmailToGroup([]string{""}, "", "", true)
	// in a testing environment, we expect these to be blank, and a nil error
	assert.Equal(t, "", resp)
	assert.Equal(t, "", id)
	assert.Nil(t, err)

	body := GenerateEmail(0, "Test")
	assert.NotEqual(t, "", body)
}
