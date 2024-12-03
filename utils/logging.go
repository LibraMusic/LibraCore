package utils

import (
	"github.com/charmbracelet/log"
)

func SetupLogger(logFormat string, logLevel log.Level) {
	if logFormat == "json" {
		log.SetFormatter(log.JSONFormatter)
	} else if logFormat == "logfmt" {
		log.SetFormatter(log.LogfmtFormatter)
	} else if logFormat != "text" {
		log.Warn("Unknown log format, defaulting to text")
	}

	log.SetLevel(logLevel)
}
