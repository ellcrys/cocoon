package config

import (
	"os"
	"strings"

	"github.com/op/go-logging"
)

var backend = logging.NewLogBackend(os.Stderr, "", 0)
var format = logging.MustStringFormatter(`%{color}%{time:15:04:05.000} [%{level:.3s}][%{module}] ▶ %{message}`)
var format2 = logging.MustStringFormatter(`%{message}`)
var format3 = logging.MustStringFormatter(`%{color}%{time:15:04:05.000} [%{level:.3s}][%{module}] %{message}`)

// MakeLogger creates a logger with a syslog backend if in a production env
// otherwise an os.Stderr backed is used.
func MakeLogger(module, prefix string) *logging.Logger {
	var b logging.Backend = backend
	if strings.ToLower(os.Getenv("COCOON_ENV")) == "production" {
		b, _ = logging.NewSyslogBackend(prefix)
	}
	backendFormatted := logging.NewBackendFormatter(b, format3)
	logging.SetBackend(backendFormatted)
	l := &logging.Logger{Module: module}
	l.SetBackend(logging.AddModuleLevel(backendFormatted))
	return l
}

// MakeLoggerMessageOnly creates a logger with a syslog backend if in a production env
// otherwise an os.Stderr backed is used. It uses a format that only displays the message.
func MakeLoggerMessageOnly(module, prefix string) *logging.Logger {
	var b logging.Backend = backend
	if strings.ToLower(os.Getenv("COCOON_ENV")) == "production" {
		b, _ = logging.NewSyslogBackend(prefix)
	}
	backendFormatted := logging.NewBackendFormatter(b, format2)
	logging.SetBackend(backendFormatted)
	l := &logging.Logger{Module: module}
	l.SetBackend(logging.AddModuleLevel(backendFormatted))
	return l
}

// MessageOnlyBackend represents a log backend that logs only the message (%{message})
var MessageOnlyBackend = logging.AddModuleLevel(logging.NewBackendFormatter(backend, format2))

// ConfigureLogger sets up the logger and it's backends
func ConfigureLogger() {
	backendFormatted := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFormatted)
}
