package config

import (
	"os"

	"github.com/op/go-logging"
)

var backend = logging.NewLogBackend(os.Stderr, "", 0)
var format = logging.MustStringFormatter(`%{color}%{time:15:04:05.000} [%{level:.3s}][%{module}] ▶ %{message}`)
var format2 = logging.MustStringFormatter(`%{message}`)
var format3 = logging.MustStringFormatter(`%{color}%{time:15:04:05.000} [%{level:.3s}][%{module}] %{message}`)

// MessageOnlyBackend represents a log backend that logs only the message (%{message})
var MessageOnlyBackend = logging.AddModuleLevel(logging.NewBackendFormatter(backend, format2))

func makeLogger(module string, format logging.Formatter) *logging.Logger {
	backendFormatted := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFormatted)
	l := &logging.Logger{Module: module}
	l.SetBackend(logging.AddModuleLevel(backendFormatted))
	return l
}

// MakeLogger creates a regular logger
func MakeLogger(module string) *logging.Logger {
	return makeLogger(module, format3)
}

// MakeLoggerMessageOnly creates a logger that displays format
func MakeLoggerMessageOnly(module string) *logging.Logger {
	return makeLogger(module, format2)
}

// ConfigureLogger sets up the logger and it's backends
func ConfigureLogger() {
	backendFormatted := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFormatted)
}
