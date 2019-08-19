package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizer(t *testing.T) {
	ConfigSetup()
	bad1 := `Hello <STYLE>.XSS{background-image:url("javascript:alert('XSS')");}</STYLE><A CLASS=XSS></A>World`
	bad2 := `<a href="javascript:alert('XSS1')" onmouseover="alert('XSS2')">XSS<a>`
	markdown := `#Hello
	
	## This is cool
	
	*strong* and _emph_
	
	(A Link)[https://www.google.com]`

	expected1 := "Hello World"
	expected2 := "XSS"
	expectedMarkdown := markdown

	found1, err := sanitize(bad1)
	assert.Nil(t, err)
	found2, err := sanitize(bad2)
	assert.Nil(t, err)
	foundMD, err := sanitize(markdown)
	assert.Nil(t, err)

	assert.Equal(t, expected1, found1)
	assert.Equal(t, expected2, found2)
	assert.Equal(t, expectedMarkdown, foundMD)
}
