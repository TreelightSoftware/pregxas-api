package api

import (
	"github.com/microcosm-cc/bluemonday"
)

var _sanitizer *bluemonday.Policy

func init() {
	_sanitizer = bluemonday.StrictPolicy()
}

func sanitize(input string) (string, error) {
	clean := _sanitizer.Sanitize(input)
	return clean, nil
}
