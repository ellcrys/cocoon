package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "0.5.0"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the API version information",
	Long:  `Show the API version information`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(`API version`, version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
