package updater

import (
	"os"
	"testing"
)

func TestYamlFileUpdater_UpdateFile(t *testing.T) {
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
			name:           "Update single package",
			fileContent:    "# depup package=test-pkg\nversion: 1.0.0\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "# depup package=test-pkg\nversion: 2.0.0\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "Update with double quoted version",
			fileContent:    "# depup package=test-pkg\nversion: \"1.0.0\"\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "# depup package=test-pkg\nversion: \"2.0.0\"\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "Update with single quotes",
			fileContent:    "# depup package=test-pkg\nversion: '1.0.0'\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "# depup package=test-pkg\nversion: '2.0.0'\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "No matching package",
			fileContent:    "# depup package=test-pkg\nversion: 1.0.0\n",
			packages:       []Package{{Name: "other-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "# depup package=test-pkg\nversion: 1.0.0\n",
			expectUpdated:  false,
			expectError:    false,
		},
		{
			name:           "Multiple packages, one update",
			fileContent:    "# depup package=pkg1\nversion: 1.0.0\n# depup package=pkg2\nversion: 1.0.0\n",
			packages:       []Package{{Name: "pkg2", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "# depup package=pkg1\nversion: 1.0.0\n# depup package=pkg2\nversion: 2.0.0\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "No depup comments",
			fileContent:    "version: 1.0.0\nother: value\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "version: 1.0.0\nother: value\n",
			expectUpdated:  false,
			expectError:    false,
		},
		{
			name:           "Dry run mode",
			fileContent:    "# depup package=test-pkg\nversion: 1.0.0\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: true},
			expectedOutput: "# depup package=test-pkg\nversion: 2.0.0\n",
			expectUpdated:  true,
			expectError:    false,
		},
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
			name:           "Different comment spacing",
			fileContent:    "#depup package=test-pkg\nversion: 1.0.0\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "#depup package=test-pkg\nversion: 2.0.0\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "Complex version number",
			fileContent:    "# depup package=test-pkg\nversion: 1.0.0-alpha.1+build.2\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0-beta.1"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "# depup package=test-pkg\nversion: 2.0.0-beta.1\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "Indented version line",
			fileContent:    "# depup package=test-pkg\n  version: 1.0.0\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "# depup package=test-pkg\n  version: 2.0.0\n",
			expectUpdated:  true,
			expectError:    false,
		},
		// --- Extended SemVer cases ---
		{
			name:           "Update with pre-release version",
			fileContent:    "# depup package=test-pkg\nversion: 1.2.3-rc.2\n",
			packages:       []Package{{Name: "test-pkg", Version: "1.2.4-beta.1"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "# depup package=test-pkg\nversion: 1.2.4-beta.1\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "Update with build metadata",
			fileContent:    "# depup package=test-pkg\nversion: \"1.2.3\"\n",
			packages:       []Package{{Name: "test-pkg", Version: "1.2.4+build.42"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "# depup package=test-pkg\nversion: \"1.2.4+build.42\"\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "Update with pre-release and build metadata",
			fileContent:    "# depup package=test-pkg\nversion: '1.2.3+build.45'\n",
			packages:       []Package{{Name: "test-pkg", Version: "1.2.4-alpha.1+meta.99"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "# depup package=test-pkg\nversion: '1.2.4-alpha.1+meta.99'\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "Indented version line with prerelease",
			fileContent:    "# depup package=test-pkg\n  version: 1.0.0\n",
			packages:       []Package{{Name: "test-pkg", Version: "1.0.1-rc.1"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "# depup package=test-pkg\n  version: 1.0.1-rc.1\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "Already matching version with prerelease",
			fileContent:    "# depup package=test-pkg\nversion: \"1.2.3-beta.1\"\n",
			packages:       []Package{{Name: "test-pkg", Version: "1.2.3-beta.1"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "# depup package=test-pkg\nversion: \"1.2.3-beta.1\"\n",
			expectUpdated:  false,
			expectError:    false,
		},
		{
			name:           "Wrong package name with prerelease",
			fileContent:    "# depup package=not-this-one\nversion: 0.0.1\n",
			packages:       []Package{{Name: "other-pkg", Version: "9.9.9-beta.9"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "# depup package=not-this-one\nversion: 0.0.1\n",
			expectUpdated:  false,
			expectError:    false,
		},
		{
			name:           "Malformed version should not match",
			fileContent:    "# depup package=test-pkg\nversion: 01.2.3\n",
			packages:       []Package{{Name: "test-pkg", Version: "1.2.3"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "# depup package=test-pkg\nversion: 01.2.3\n",
			expectUpdated:  false,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with test content
			tempFile, err := createTempFileWithContent(tt.fileContent, ".yaml")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile)

			// Create updater
			updater := NewYamlFileUpdater()

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

func TestYamlFileUpdater_InlineComments(t *testing.T) {
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
			name:           "Basic inline comment",
			fileContent:    "version: 1.0.0 # depup package=test-pkg\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "version: 2.0.0 # depup package=test-pkg\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "Inline comment with double quotes",
			fileContent:    "version: \"1.0.0\" # depup package=test-pkg\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "version: \"2.0.0\" # depup package=test-pkg\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "Inline comment with single quotes",
			fileContent:    "version: '1.0.0' # depup package=test-pkg\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "version: '2.0.0' # depup package=test-pkg\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "Inline comment with no matching package",
			fileContent:    "version: 1.0.0 # depup package=test-pkg\n",
			packages:       []Package{{Name: "other-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "version: 1.0.0 # depup package=test-pkg\n",
			expectUpdated:  false,
			expectError:    false,
		},
		{
			name:           "Inline comment with indented line",
			fileContent:    "  version: 1.0.0 # depup package=test-pkg\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "  version: 2.0.0 # depup package=test-pkg\n",
			expectUpdated:  true,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with test content
			tempFile, err := createTempFileWithContent(tt.fileContent, ".yaml")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile)

			// Create updater
			updater := NewYamlFileUpdater()

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

func TestYamlFileUpdater_KubernetesFiles(t *testing.T) {
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
			name: "Kubernetes deployment with depup comment",
			fileContent: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  template:
    spec:
      containers:
      # depup package=my-app
      - image: company/my-app:1.0.0
`,
			packages: []Package{{Name: "my-app", Version: "2.0.0"}},
			options:  FileUpdaterOptions{DryRun: false},
			expectedOutput: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  template:
    spec:
      containers:
      # depup package=my-app
      - image: company/my-app:2.0.0
`,
			expectUpdated: true,
			expectError:   false,
		},
		{
			name: "Kubernetes deployment with no matching package",
			fileContent: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  template:
    spec:
      containers:
      # depup package=my-app
      - image: company/my-app:1.0.0
`,
			packages: []Package{{Name: "different-app", Version: "2.0.0"}},
			options:  FileUpdaterOptions{DryRun: false},
			expectedOutput: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  template:
    spec:
      containers:
      # depup package=my-app
      - image: company/my-app:1.0.0
`,
			expectUpdated: false,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with test content
			tempFile, err := createTempFileWithContent(tt.fileContent, ".yaml")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile)

			// Create updater
			updater := NewYamlFileUpdater()

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

func TestYamlFileUpdater_Supports(t *testing.T) {
	updater := NewYamlFileUpdater()

	tests := []struct {
		extension string
		expected  bool
	}{
		{".yaml", true},
		{".yml", true},
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
