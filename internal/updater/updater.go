package updater

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var /* const */ namePattern = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
var /* const */ versionPattern = regexp.MustCompile(`(["']?)(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?(["']?)`)

// Package represents a dependency package with a name and version
// to be updated in configuration files
type Package struct {
	Name    string // Name of the package identifier
	Version string // Version of the package (semantic version format)
}

func (p *Package) String() string {
	return fmt.Sprintf("%s=%s", p.Name, p.Version)
}

func (p *Package) Validate() error {
	var errs []error

	if !versionPattern.MatchString(p.Version) {
		errs = append(errs, fmt.Errorf("invalid version format: %s", p.Version))
	}
	if !namePattern.MatchString(p.Name) {
		errs = append(errs, fmt.Errorf("invalid name format: %s", p.Name))
	}
	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

// FileUpdaterOptions contains configuration for file update operations
type FileUpdaterOptions struct {
	DryRun bool // When true, changes are not written to files
}

// FileUpdater is an interface that defines the behavior of a concrete updater
// Implementations handle different file formats (yaml, json, etc.)
type FileUpdater interface {
	// Supports checks if the updater supports the given file extension
	Supports(fileExtension string) bool

	// GetSupportedExtensions returns a list of file extensions supported by the updater
	GetSupportedExtensions() []string

	// UpdateFile updates the dependencies in the specified file
	// Returns the updated content, whether changes were made, and any error
	UpdateFile(filePath string, packages []Package, options FileUpdaterOptions) (string, bool, error)
}

// Option represents a function that configures the Updater
// Uses the functional options pattern for flexible configuration
type Option func(*Updater)

// WithDryRun configures the updater to run in dry-run mode (no changes applied)
// When enabled, changes are only simulated but not written to disk
func WithDryRun(dryRun bool) Option {
	return func(u *Updater) {
		u.dryRun = dryRun
	}
}

// WithRecursive configures the updater to scan directories recursively
// When enabled, subdirectories are traversed when processing a directory
func WithRecursive(recursive bool) Option {
	return func(u *Updater) {
		u.recursive = recursive
	}
}

// WithFileExtensions specifies which file extensions to process
// This restricts the updater to only handle files with the specified extensions
func WithFileExtensions(extensions []string) Option {
	return func(u *Updater) {
		u.fileExtensions = extensions
	}
}

// Updater is the main struct that orchestrates the dependency update process
// It manages file discovery and delegates actual updates to specialized implementations
type Updater struct {
	// updaters is a slice of FileUpdater implementations for different file types
	updaters []FileUpdater

	// configuration options
	dryRun         bool     // When true, changes are not written to files
	recursive      bool     // When true, subdirectories are processed
	fileExtensions []string // List of file extensions to consider for updates
}

// NewUpdater creates a new instance of the Updater with the provided options
// Default configuration includes YAML support and common settings
func NewUpdater(options ...Option) *Updater {
	u := &Updater{
		updaters: []FileUpdater{
			NewYamlFileUpdater(),
			NewHclFileUpdater(),
			NewDotEnvFileUpdater(),
		},
		// Default values
		dryRun:         false,
		recursive:      true,
		fileExtensions: []string{},
	}

	for _, updater := range u.updaters {
		u.fileExtensions = append(u.fileExtensions, updater.GetSupportedExtensions()...)
	}

	// Apply all provided options to override defaults
	for _, opt := range options {
		opt(u)
	}

	return u
}

// Update processes the entrypoint (file or directory) and updates dependencies
// based on the provided packages list and configuration options
func (u *Updater) Update(entrypoint string, packages []Package) error {
	var errs []error
	for _, pkg := range packages {
		if err := pkg.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("invalid package %s: %w", pkg, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid packages: %w", errors.Join(errs...))
	}

	// Verify the entrypoint exists
	fileInfo, err := os.Stat(entrypoint)
	if err != nil {
		return err
	}

	// Convert to absolute path for consistency in error messages and processing
	entrypoint, err = filepath.Abs(entrypoint)
	if err != nil {
		return err
	}

	// Prepare options for file updaters
	updaterOptions := FileUpdaterOptions{
		DryRun: u.dryRun,
	}

	// Handle single file case
	if !fileInfo.IsDir() {
		return u.processFile(entrypoint, packages, updaterOptions)
	}

	// Handle directory case
	if fileInfo.IsDir() && !u.recursive {
		// Process only files in the top-level directory when recursive is false
		files, err := os.ReadDir(entrypoint)
		if err != nil {
			return err
		}

		for _, file := range files {
			if !file.IsDir() {
				ext := filepath.Ext(file.Name())
				for _, allowedExt := range u.fileExtensions {
					if ext == allowedExt {
						err = u.processFile(filepath.Join(entrypoint, file.Name()), packages, updaterOptions)
						if err != nil {
							return err
						}
						break
					}
				}
			}
		}
	} else if u.recursive {
		// Process all files recursively when recursive flag is true
		err = filepath.Walk(entrypoint, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				ext := filepath.Ext(path)
				for _, allowedExt := range u.fileExtensions {
					if ext == allowedExt {
						err = u.processFile(path, packages, updaterOptions)
						if err != nil {
							return err
						}
						break
					}
				}
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// isFileExtensionSupported checks if the file extension is in the configured extensions list
// Returns true if the file should be processed, false otherwise
func (u *Updater) isFileExtensionSupported(filePath string) bool {
	fileExtension := strings.ToLower(filepath.Ext(filePath))
	fileName := filepath.Base(filePath)

	for _, pattern := range u.fileExtensions {
		// First check exact extension match
		if pattern == fileExtension {
			return true
		}

		// Then check for glob pattern match
		if strings.Contains(pattern, "*") || strings.Contains(pattern, "?") || strings.Contains(pattern, "[") {
			matched, err := filepath.Match(pattern, fileName)
			if err == nil && matched {
				return true
			}
		}
	}
	return false
}

// getFileUpdater returns the appropriate FileUpdater for a given file extension
// Returns an error if no suitable updater is found
func (u *Updater) getFileUpdater(fileExtension string) (FileUpdater, error) {
	for _, updater := range u.updaters {
		if updater.Supports(fileExtension) {
			return updater, nil
		}
	}
	return nil, fmt.Errorf("no updater found for file extension: %s", fileExtension)
}

// processFile handles updating a single file with the provided packages
// Selects the appropriate updater based on file extension and delegates the actual update
func (u *Updater) processFile(filePath string, packages []Package, options FileUpdaterOptions) error {
	// Skip files with unsupported extensions
	if !u.isFileExtensionSupported(filePath) {
		return nil // Exit silently if file extension is not supported
	}

	// Get the appropriate updater for this file type
	updater, err := u.getFileUpdater(filepath.Ext(filePath))
	if err != nil {
		return fmt.Errorf("no updater found for file extension: %s", filepath.Ext(filePath))
	}

	// Perform the update operation
	updatedContent, hasBeenUpdated, err := updater.UpdateFile(filePath, packages, options)
	if err != nil {
		return err
	}

	// In dry-run mode, output what would change instead of modifying files
	if u.dryRun && hasBeenUpdated {
		fmt.Printf("Dry run mode - updated content for %s:\n%s\n", filePath, updatedContent)
	}

	return nil
}
