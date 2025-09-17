package libracore

import "github.com/charmbracelet/log"

func SetupLogger(logFormat string, logLevel log.Level) {
	log.SetLevel(logLevel)

	switch logFormat {
	case "json":
		log.SetFormatter(log.JSONFormatter)
	case "logfmt":
		log.SetFormatter(log.LogfmtFormatter)
	case "text":
		log.SetFormatter(log.TextFormatter)
	default:
		log.SetFormatter(log.TextFormatter)
		log.Warn("Unknown log format, defaulting to text")
	}
}
