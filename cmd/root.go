package cmd

import (
	"github.com/spf13/cobra"
)

var (
	// AppName for 应用名称
	AppName string
	// AppVersion for 应用版本
	AppVersion string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:              "arc-consumer",
	TraverseChildren: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.AddCommand(serverCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

}
