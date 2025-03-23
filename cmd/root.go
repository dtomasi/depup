package cmd

import (
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
// All other commands are added as subcommands to this root command.
var rootCmd = &cobra.Command{
	Use:   "depup",
	Short: "A tool for dependency management", // Displayed in help output
	Long: `Depup is a CLI tool that helps manage and update dependencies
in your projects efficiently and reliably.`, // Detailed description for help
	// No Run function as this command serves as a container for subcommands
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This function is called by main.main(). It only needs to happen once.
func Execute() error {
	// Execute will run the command and return any errors
	return rootCmd.Execute()
}

func init() {
	// init is called when the package is imported
	// Use this function to configure the root command's persistent flags
	// Persistent flags are inherited by all subcommands
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file path")
	// StringP defines a flag with a string value and a short flag alternative
	// The arguments are: name, shorthand, default value, and usage/description
}
