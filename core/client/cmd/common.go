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
