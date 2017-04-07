package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// sigRmCmd represents the sig-rm command
var sigRmCmd = &cobra.Command{
	Use:   "sig-rm [OPTIONS] COCOON IDENTITY [IDENTITY...]",
	Short: "Remove one or more signatories from a cocoon",
	Long:  `Remove one or more signatories from a cocoon`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		fmt.Println("sig-rm called")
	},
}

func init() {
	RootCmd.AddCommand(sigRmCmd)
}
