package util

import (
	"os"

	"github.com/Sirupsen/logrus"
)

// Log is a global logger
var Log = logrus.New()

func init() {
	Log.Out = os.Stderr
	Log.Formatter = &logrus.TextFormatter{}
}

// LogSetFormatter updates the formatter of the global logger
func LogSetFormatter(f logrus.Formatter) {
	Log.Formatter = f
}

// LogHook attaches an hook to the global logger
func LogHook(hook logrus.Hook) {
	Log.Hooks.Add(hook)
}

// Println prints errors to stderr using a looger. Uses Log.Error if
// an error is passed, otherwise uses info.
func Println(args ...interface{}) {
	if len(args) > 0 {
		switch args[0].(type) {
		case error:
			Log.Error(args...)
		default:
			Log.Info(args...)
		}
	}
}
