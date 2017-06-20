package cmd

import (
	"fmt"

	"github.com/ellcrys/cocoon/core/client/client"
	"github.com/spf13/cobra"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display system-wide information",
	Long:  `Display system-wide information`,
	Run: func(cmd *cobra.Command, args []string) {

		userSession, err := client.GetUserSessionToken()
		if err != nil {
			return
		}

		fmt.Println(`User:
  Email: ` + userSession.Email + `

Client:
  Version:     0.5.0
  API version: 0.5.0
  Go version:  go1.8
  
Core:
  API:         https://rpc.ellcrys.co
  Version:     0.5.0
  API version: 0.5.0
  Go version:  go1.8
		`)
	},
}

func init() {
	RootCmd.AddCommand(infoCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// infoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// infoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
