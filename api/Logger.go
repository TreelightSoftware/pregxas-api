package api

import (
	"strings"

	"github.com/sirupsen/logrus"
)

// Log provides an abstract way to log data
func Log(level, message, code string, otherData map[string]string) {
	fields := logrus.Fields{}
	for entry := range otherData {
		fields[entry] = otherData[entry]
	}

	if code != "" {
		fields["code"] = code
	}
	level = strings.ToLower(level)
	switch level {
	case "info":
		Config.Logger.WithFields(fields).Info(message)
	case "warning":
		Config.Logger.WithFields(fields).Warning(message)
	case "error":
		Config.Logger.WithFields(fields).Error(message)
	}

}
