package cmd

import (
	"os"

	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// UsageError  displays Usage error
func UsageError(l *logging.Logger, cmd *cobra.Command, msg, helpCmd string) {
	l.Info(msg)
	l.Infof("See '%s'", helpCmd)
	l.Info("")
	l.Info("Usage:", cmd.Root().Name(), cmd.Use)
	l.Info("")
	l.Info(cmd.Long)
	os.Exit(1)
}

// UsageError2 displays simple Usage error
func UsageError2(l *logging.Logger, cmd *cobra.Command, msg, helpCmd string) {
	l.Info(msg)
	l.Infof("See '%s'", helpCmd)
	os.Exit(1)
}

func toStringOr(v interface{}, def string) string {
	if v == nil {
		return def
	}
	return v.(string)
}

func toIntOr(v interface{}, def int) int {
	if v == nil {
		return def
	}
	return v.(int)
}

func toInt32Or(v interface{}, def int32) int32 {
	if v == nil {
		return def
	}
	return v.(int32)
}
