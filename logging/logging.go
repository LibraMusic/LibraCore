package logging

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/DevReaper0/libra/config"
)

func Init() {
	if config.Conf.Logs.Debug {
		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		output.FormatLevel = func(i interface{}) string {
			return "| " + colorize(strings.ToUpper(fmt.Sprintf("%-5s", i)), levelColors[fmt.Sprintf("%s", i)]) + " |"
		}
		log.Logger = log.Output(output)

		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

func Debug() *zerolog.Event {
	return log.Debug()
}

func Info() *zerolog.Event {
	return log.Info()
}

func Warn() *zerolog.Event {
	if config.Conf.Logs.ErrorWarnings {
		return Error()
	}
	return log.Warn()
}

func Error() *zerolog.Event {
	if config.Conf.Logs.AllErrorsFatal {
		return Fatal()
	}
	return log.Error()
}

func Fatal() *zerolog.Event {
	return log.Fatal()
}

const (
	colorBlack = iota + 30
	colorRed
	colorGreen
	colorYellow
	colorBlue
	colorMagenta
	colorCyan
	colorWhite
)

var levelColors = map[string]int{
	"trace": colorBlue,
	"debug": 0,
	"info":  colorGreen,
	"warn":  colorYellow,
	"error": colorRed,
	"fatal": colorRed,
	"panic": colorRed,
}

// colorize is based on colorize from zerolog (https://github.com/rs/zerolog/blob/6abadab4881e4af6d122332f6aef0365507c248a/console.go#L374)
// colorize returns the string s wrapped in ANSI code c, unless disabled is true or c is 0.
func colorize(s interface{}, c int) string {
	disabled := false

	e := os.Getenv("NO_COLOR")
	if e != "" || c == 0 {
		disabled = true
	}

	if disabled {
		return fmt.Sprintf("%s", s)
	}
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, s)
}
