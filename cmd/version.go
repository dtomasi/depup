package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// These variables are meant to be set during build using ldflags
	// Example: go build -ldflags "-X github.com/dtomasi/depup/cmd.version=1.0.0"
	version = "dev"     // Semantic version of the application
	commit  = "none"    // Git commit hash
	date    = "unknown" // Build timestamp
)

// versionCmd represents the version command that prints version information
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of depup",
	// Run defines the command's behavior
	Run: func(cmd *cobra.Command, args []string) {
		if short, _ := cmd.Flags().GetBool("short"); short {
			// Print only the version number
			fmt.Println(version)

			return
		}

		// Print version information in a standardized format
		fmt.Printf("depup version %s (commit: %s, built at: %s)\n", version, commit, date)
	},
}

func init() {
	// Register the version command as a subcommand of the root command
	rootCmd.AddCommand(versionCmd)

	// Flag to print only the version number
	versionCmd.Flags().BoolP("short", "s", false, "Print only the version number")
}
