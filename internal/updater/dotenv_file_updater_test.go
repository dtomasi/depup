package updater

import (
	"os"
	"testing"
)

func TestDotEnvFileUpdater_UpdateFile(t *testing.T) {
	tests := []struct {
		name           string
		fileContent    string
		packages       []Package
		options        FileUpdaterOptions
		expectedOutput string
		expectUpdated  bool
		expectError    bool
	}{
		{
			name:           "Empty file",
			fileContent:    "",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "",
			expectUpdated:  false,
			expectError:    false,
		},
		{
			name:           "File with no depup comments",
			fileContent:    "VERSION=1.0.0\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "VERSION=1.0.0\n",
			expectUpdated:  false,
			expectError:    false,
		},
		{
			name:           "Previous line comment with matching package",
			fileContent:    "# depup package=test-pkg\nVERSION=1.0.0\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "# depup package=test-pkg\nVERSION=2.0.0\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "Inline comment with matching package",
			fileContent:    "VERSION=1.0.0 # depup package=test-pkg\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "VERSION=2.0.0 # depup package=test-pkg\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "Previous line comment with double quoted value",
			fileContent:    "# depup package=test-pkg\nVERSION=\"1.0.0\"\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "# depup package=test-pkg\nVERSION=\"2.0.0\"\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "Inline comment with single quoted value",
			fileContent:    "VERSION='1.0.0' # depup package=test-pkg\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "VERSION='2.0.0' # depup package=test-pkg\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "Multiple packages, one update",
			fileContent:    "# depup package=pkg1\nPKG1_VERSION=1.0.0\n\n# depup package=pkg2\nPKG2_VERSION=1.0.0\n",
			packages:       []Package{{Name: "pkg2", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "# depup package=pkg1\nPKG1_VERSION=1.0.0\n\n# depup package=pkg2\nPKG2_VERSION=2.0.0\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "No matching package name",
			fileContent:    "# depup package=test-pkg\nVERSION=1.0.0\n",
			packages:       []Package{{Name: "other-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "# depup package=test-pkg\nVERSION=1.0.0\n",
			expectUpdated:  false,
			expectError:    false,
		},
		{
			name:           "No version change needed",
			fileContent:    "VERSION=1.0.0 # depup package=test-pkg\n",
			packages:       []Package{{Name: "test-pkg", Version: "1.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "VERSION=1.0.0 # depup package=test-pkg\n",
			expectUpdated:  false,
			expectError:    false,
		},
		{
			name:           "Dry run mode",
			fileContent:    "# depup package=test-pkg\nVERSION=1.0.0\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: true},
			expectedOutput: "# depup package=test-pkg\nVERSION=2.0.0\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "Indented line",
			fileContent:    "# depup package=test-pkg\n  VERSION=1.0.0\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "# depup package=test-pkg\n  VERSION=2.0.0\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "With prerelease version",
			fileContent:    "# depup package=test-pkg\nVERSION=1.0.0\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0-beta.1"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "# depup package=test-pkg\nVERSION=2.0.0-beta.1\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "With build metadata",
			fileContent:    "VERSION=\"1.0.0\" # depup package=test-pkg\n",
			packages:       []Package{{Name: "test-pkg", Version: "1.0.0+build.123"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "VERSION=\"1.0.0+build.123\" # depup package=test-pkg\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "With prerelease and build metadata",
			fileContent:    "# depup package=test-pkg\nVERSION='1.0.0'\n",
			packages:       []Package{{Name: "test-pkg", Version: "1.0.0-alpha.1+build.123"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "# depup package=test-pkg\nVERSION='1.0.0-alpha.1+build.123'\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "Example from dotenv/.env file - previous line comment",
			fileContent:    "# depup package=postgres\nPOSTGRES_VERSION=1.2.3\n",
			packages:       []Package{{Name: "postgres", Version: "1.3.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "# depup package=postgres\nPOSTGRES_VERSION=1.3.0\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "Example from dotenv/.env file - inline comment",
			fileContent:    "REDIS_VERSION=4.0.0 # depup package=redis\n",
			packages:       []Package{{Name: "redis", Version: "4.2.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "REDIS_VERSION=4.2.0 # depup package=redis\n",
			expectUpdated:  true,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with test content
			tempFile, err := createTempFileWithContent(tt.fileContent, ".env")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile)

			// Create updater
			updater := NewDotEnvFileUpdater()

			// Call the method
			output, updated, err := updater.UpdateFile(tempFile, tt.packages, tt.options)

			// Check error expectation
			if (err != nil) != tt.expectError {
				t.Errorf("UpdateFile() error = %v, expectError %v", err, tt.expectError)
				return
			}

			// Check updated flag
			if updated != tt.expectUpdated {
				t.Errorf("UpdateFile() updated = %v, expectUpdated %v", updated, tt.expectUpdated)
			}

			// Check output content
			if output != tt.expectedOutput {
				t.Errorf("UpdateFile() output = %q, expectedOutput %q", output, tt.expectedOutput)
			}

			// If not dry run and expected update, check file contents
			if !tt.options.DryRun && tt.expectUpdated {
				content, err := os.ReadFile(tempFile)
				if err != nil {
					t.Errorf("Failed to read temp file: %v", err)
					return
				}

				if string(content) != tt.expectedOutput {
					t.Errorf("File content = %q, expectedOutput %q", string(content), tt.expectedOutput)
				}
			}
		})
	}
}

func TestDotEnvFileUpdater_ComplexCases(t *testing.T) {
	tests := []struct {
		name           string
		fileContent    string
		packages       []Package
		options        FileUpdaterOptions
		expectedOutput string
		expectUpdated  bool
		expectError    bool
	}{
		{
			name: "Mixed comment styles and values",
			fileContent: "# Regular comment\n" +
				"APP_VERSION=1.0.0 # depup package=my-app\n" +
				"# depup package=database\n" +
				"DB_VERSION=\"2.1.0\"\n" +
				"CACHE_VERSION='1.5.0' # depup package=cache\n",
			packages: []Package{
				{Name: "my-app", Version: "1.1.0"},
				{Name: "database", Version: "2.2.0"},
				{Name: "cache", Version: "1.6.0"},
			},
			options: FileUpdaterOptions{DryRun: false},
			expectedOutput: "# Regular comment\n" +
				"APP_VERSION=1.1.0 # depup package=my-app\n" +
				"# depup package=database\n" +
				"DB_VERSION=\"2.2.0\"\n" +
				"CACHE_VERSION='1.6.0' # depup package=cache\n",
			expectUpdated: true,
			expectError:   false,
		},
		{
			name: "Environment variables with spaces",
			fileContent: "# depup package=app\n" +
				"APP_VERSION = 1.0.0\n" +
				"OTHER_VAR = some value # depup package=other\n",
			packages: []Package{
				{Name: "app", Version: "2.0.0"},
				{Name: "other", Version: "3.0.0"},
			},
			options: FileUpdaterOptions{DryRun: false},
			expectedOutput: "# depup package=app\n" +
				"APP_VERSION = 2.0.0\n" +
				"OTHER_VAR = some value # depup package=other\n",
			expectUpdated: true,
			expectError:   false,
		},
		{
			name: "Comments with no corresponding variable",
			fileContent: "# Regular comment\n" +
				"# depup package=orphaned\n" +
				"# Another comment\n" +
				"VERSION=1.0.0 # Not a depup comment\n",
			packages: []Package{
				{Name: "orphaned", Version: "9.9.9"},
			},
			options: FileUpdaterOptions{DryRun: false},
			expectedOutput: "# Regular comment\n" +
				"# depup package=orphaned\n" +
				"# Another comment\n" +
				"VERSION=1.0.0 # Not a depup comment\n",
			expectUpdated: false,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with test content
			tempFile, err := createTempFileWithContent(tt.fileContent, ".env")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile)

			// Create updater
			updater := NewDotEnvFileUpdater()

			// Call the method
			output, updated, err := updater.UpdateFile(tempFile, tt.packages, tt.options)

			// Check error expectation
			if (err != nil) != tt.expectError {
				t.Errorf("UpdateFile() error = %v, expectError %v", err, tt.expectError)
				return
			}

			// Check updated flag
			if updated != tt.expectUpdated {
				t.Errorf("UpdateFile() updated = %v, expectUpdated %v", updated, tt.expectUpdated)
			}

			// Check output content
			if output != tt.expectedOutput {
				t.Errorf("UpdateFile() output = %q, expectedOutput %q", output, tt.expectedOutput)
			}
		})
	}
}

func TestDotEnvFileUpdater_Supports(t *testing.T) {
	updater := NewDotEnvFileUpdater()

	tests := []struct {
		extension string
		expected  bool
	}{
		{".env", true},
		{".yaml", false},
		{".yml", false},
		{".hcl", false},
		{".tf", false},
		{".json", false},
		{".txt", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.extension, func(t *testing.T) {
			result := updater.Supports(tt.extension)
			if result != tt.expected {
				t.Errorf("Supports(%q) = %v, expected %v", tt.extension, result, tt.expected)
			}
		})
	}
}
