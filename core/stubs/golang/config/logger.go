package config

import (
	"os"

	"github.com/op/go-logging"
)

var backend = logging.NewLogBackend(os.Stderr, "", 0)
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} [%{level:.3s}][%{module}](%{shortfile}) ▶ %{message}`,
)

// ConfigureLogger sets up the logger and it's backends
func ConfigureLogger() {
	backendFormatted := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFormatted)
}
