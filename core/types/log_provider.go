package types

import (
	"time"

	context "golang.org/x/net/context"
)

// LogProvider provides an interface to accessing
// logs from a log storage provider
type LogProvider interface {
	Init(config map[string]interface{}) error
	Get(context context.Context, logName string, numEntries int, source string) ([]LogMessage, error)
}

// LogMessage represents a log message
type LogMessage struct {
	ID        string
	Text      string
	Timestamp time.Time
}
