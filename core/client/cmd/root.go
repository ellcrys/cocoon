package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "ellcrys",
	Short: "Ellcrys client to the blockchain-inspired, shared ledger and smart contract platform",
	Long:  `Ellcrys client to the blockchain-inspired, shared ledger and smart contract platform`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		showVersion, _ := cmd.Flags().GetBool("version")
		if showVersion {
			fmt.Println("Cocoon version 1.0.0")
			return
		}
		fmt.Println(cmd.Long)
		fmt.Println("")
		cmd.Usage()
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.Flags().BoolP("version", "v", false, "Display the Cocoon version information")
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigName(".client") // name of config file (without extension)
	viper.AddConfigPath("$HOME")   // adding home directory as first search path
	viper.AutomaticEnv()           // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
