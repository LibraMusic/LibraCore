package logging

import (
	"github.com/charmbracelet/log"

	"github.com/LibraMusic/LibraCore/config"
)

func Init() {
	if config.Conf.Logs.LogFormat == "json" {
		log.SetFormatter(log.JSONFormatter)
	} else if config.Conf.Logs.LogFormat == "logfmt" {
		log.SetFormatter(log.LogfmtFormatter)
	} else if config.Conf.Logs.LogFormat != "text" {
		log.Warn("Unknown log format, defaulting to text")
	}

	log.SetLevel(config.Conf.Logs.LogLevel)
}

func Debug(msg interface{}, keyvals ...interface{}) {
	log.Debug(msg, keyvals)
}

func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func Info(msg interface{}, keyvals ...interface{}) {
	log.Info(msg, keyvals)
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Warn(msg interface{}, keyvals ...interface{}) {
	if config.Conf.Logs.ErrorWarnings {
		Error(msg, keyvals)
	}
	log.Warn(msg, keyvals)
}

func Warnf(format string, args ...interface{}) {
	if config.Conf.Logs.ErrorWarnings {
		Errorf(format, args...)
	}
	log.Warnf(format, args...)
}

func Error(msg interface{}, keyvals ...interface{}) {
	if config.Conf.Logs.AllErrorsFatal {
		Fatal(msg, keyvals)
	}
	log.Error(msg, keyvals)
}

func Errorf(format string, args ...interface{}) {
	if config.Conf.Logs.AllErrorsFatal {
		Fatalf(format, args...)
	}
	log.Errorf(format, args...)
}

func Fatal(msg interface{}, keyvals ...interface{}) {
	log.Fatal(msg, keyvals)
}

func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

func Print(msg interface{}, keyvals ...interface{}) {
	log.Print(msg, keyvals)
}

func Printf(format string, args ...interface{}) {
	log.Printf(format, args...)
}
