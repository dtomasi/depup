package cmd

import (
	"fmt"
	"github.com/dtomasi/depup/internal/updater"
	"github.com/spf13/cobra"
	"regexp"
	"strings"
)

var packagePattern = regexp.MustCompile(`^[^@]+@v?[\d]+\.[\d]+\.[\d]+$`) // Regular expression to validate package format

// updateCmd represents the update command for updating dependencies
var updateCmd = &cobra.Command{
	Use:   "update DIR", // Command syntax showing required directory argument
	Short: "Update dependencies to their latest versions",
	Long:  `Scan and update dependencies according to specified criteria. Requires a directory path as entry point.`,
	Args:  cobra.ExactArgs(1), // Validate that exactly one argument (directory path) is provided
	// RunE allows returning an error instead of just handling it internally
	// This provides better error handling and is more testable
	RunE: func(cmd *cobra.Command, args []string) error {
		// Retrieve flag values by name
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		recursive, _ := cmd.Flags().GetBool("recursive")
		rawPackages, _ := cmd.Flags().GetStringArray("package")
		fileExtensions, _ := cmd.Flags().GetStringArray("extension")

		var packages []updater.Package
		for _, pkg := range rawPackages {
			// Validate the package format (IDENTIFIER@SEMVER_VERSION)
			if !packagePattern.MatchString(pkg) {
				return fmt.Errorf("invalid package format: %s", pkg)
			}

			// Split the package into name and version
			parts := strings.Split(pkg, "@")
			packages = append(packages, updater.Package{Name: parts[0], Version: parts[1]})
		}

		if len(packages) == 0 {
			return fmt.Errorf("no packages to update")
		}

		updater := updater.NewUpdater(
			updater.WithDryRun(dryRun),
			updater.WithRecursive(recursive),
			updater.WithFileExtensions(fileExtensions),
		)

		return updater.Update(args[0], packages)
	},
}

func init() {
	// Register the update command as a subcommand of the root command
	rootCmd.AddCommand(updateCmd)

	// Flag to specify dry-run mode for update command
	updateCmd.Flags().BoolP("dry-run", "d", false, "Show what would be updated without making changes")

	// Flag to specify recursive lookup for files in a directory
	updateCmd.Flags().BoolP("recursive", "r", false, "Make depup lookup for files recursively, if a directory is passed as argument")

	// Flag to specify packages to update in the format IDENTIFIER@SEMVER_VERSION
	updateCmd.Flags().StringArrayP("package", "p", []string{}, "Specify dependencies to update in the format IDENTIFIER@SEMVER_VERSION (-p package@1.2.3)")

	// Flag to specify file extensions to include in the search
	updateCmd.Flags().StringArrayP("extension", "e", []string{".yaml", ".yml"}, "Specify file extensions to include in the search")
}
