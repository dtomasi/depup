package updater

import (
	"os"
	"testing"
)

func TestHclFileUpdater_UpdateFile(t *testing.T) {
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
			fileContent:    "version = \"1.0.0\"\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "version = \"1.0.0\"\n",
			expectUpdated:  false,
			expectError:    false,
		},
		{
			name:           "Inline hash comment with matching package",
			fileContent:    "version = \"1.0.0\" # depup package=test-pkg\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "version = \"2.0.0\" # depup package=test-pkg\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "Inline slash comment with matching package",
			fileContent:    "version = \"1.0.0\" // depup package=test-pkg\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "version = \"2.0.0\" // depup package=test-pkg\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "Previous line hash comment with matching package",
			fileContent:    "# depup package=test-pkg\nversion = \"1.0.0\"\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "# depup package=test-pkg\nversion = \"2.0.0\"\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "Previous line slash comment with matching package",
			fileContent:    "// depup package=test-pkg\nversion = \"1.0.0\"\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "// depup package=test-pkg\nversion = \"2.0.0\"\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "Inline comment with no matching package",
			fileContent:    "version = \"1.0.0\" # depup package=test-pkg\n",
			packages:       []Package{{Name: "other-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "version = \"1.0.0\" # depup package=test-pkg\n",
			expectUpdated:  false,
			expectError:    false,
		},
		{
			name: "Multiple lines with different packages",
			fileContent: "// depup package=pkg1\nversion_pkg1 = \"1.0.0\"\n" +
				"version_other = \"3.0.0\" # Not a depup comment\n" +
				"# depup package=pkg2\nversion_pkg2 = \"1.5.0\"\n",
			packages: []Package{
				{Name: "pkg1", Version: "2.0.0"},
				{Name: "pkg2", Version: "2.5.0"},
			},
			options: FileUpdaterOptions{DryRun: false},
			expectedOutput: "// depup package=pkg1\nversion_pkg1 = \"2.0.0\"\n" +
				"version_other = \"3.0.0\" # Not a depup comment\n" +
				"# depup package=pkg2\nversion_pkg2 = \"2.5.0\"\n",
			expectUpdated: true,
			expectError:   false,
		},
		{
			name:           "Dry run mode should not update file",
			fileContent:    "version = \"1.0.0\" # depup package=test-pkg\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: true},
			expectedOutput: "version = \"2.0.0\" # depup package=test-pkg\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "Different quote styles",
			fileContent:    "version = '1.0.0' # depup package=test-pkg\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "version = '2.0.0' # depup package=test-pkg\n",
			expectUpdated:  true,
			expectError:    false,
		},
		{
			name:           "No version change needed",
			fileContent:    "version = \"1.0.0\" # depup package=test-pkg\n",
			packages:       []Package{{Name: "test-pkg", Version: "1.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "version = \"1.0.0\" # depup package=test-pkg\n",
			expectUpdated:  false,
			expectError:    false,
		},
		{
			name:           "Inline comment with indented line",
			fileContent:    "  version = \"1.0.0\" # depup package=test-pkg\n",
			packages:       []Package{{Name: "test-pkg", Version: "2.0.0"}},
			options:        FileUpdaterOptions{DryRun: false},
			expectedOutput: "  version = \"2.0.0\" # depup package=test-pkg\n",
			expectUpdated:  true,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with test content
			tempFile, err := createTempFileWithContent(tt.fileContent, ".hcl")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile)

			// Create updater
			updater := NewHclFileUpdater()

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

func TestHclFileUpdater_TerraformFiles(t *testing.T) {
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
			name: "Terraform provider version",
			fileContent: `terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      # depup package=aws-provider
      version = "4.0.0"
    }
  }
}
`,
			packages: []Package{{Name: "aws-provider", Version: "4.5.0"}},
			options:  FileUpdaterOptions{DryRun: false},
			expectedOutput: `terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      # depup package=aws-provider
      version = "4.5.0"
    }
  }
}
`,
			expectUpdated: true,
			expectError:   false,
		},
		{
			name: "Terraform module version",
			fileContent: `module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  // depup package=aws-vpc-module
  version = "3.14.0"

  name = "my-vpc"
  cidr = "10.0.0.0/16"
}
`,
			packages: []Package{{Name: "aws-vpc-module", Version: "3.19.0"}},
			options:  FileUpdaterOptions{DryRun: false},
			expectedOutput: `module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  // depup package=aws-vpc-module
  version = "3.19.0"

  name = "my-vpc"
  cidr = "10.0.0.0/16"
}
`,
			expectUpdated: true,
			expectError:   false,
		},
		{
			name: "Terraform with multiple versions",
			fileContent: `terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      # depup package=aws-provider
      version = "4.0.0"
    }

    azurerm = {
      source  = "hashicorp/azurerm"
      // depup package=azure-provider
      version = "3.0.0"
    }
  }
}
`,
			packages: []Package{
				{Name: "aws-provider", Version: "4.5.0"},
				{Name: "azure-provider", Version: "3.2.0"},
			},
			options: FileUpdaterOptions{DryRun: false},
			expectedOutput: `terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      # depup package=aws-provider
      version = "4.5.0"
    }

    azurerm = {
      source  = "hashicorp/azurerm"
      // depup package=azure-provider
      version = "3.2.0"
    }
  }
}
`,
			expectUpdated: true,
			expectError:   false,
		},
		{
			name: "Terraform with version constraint",
			fileContent: `terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      # depup package=aws-provider
      version = ">= 4.0.0"
    }
  }
}
`,
			packages: []Package{{Name: "aws-provider", Version: "4.5.0"}},
			options:  FileUpdaterOptions{DryRun: false},
			expectedOutput: `terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      # depup package=aws-provider
      version = ">= 4.5.0"
    }
  }
}
`,
			expectUpdated: true,
			expectError:   false,
		},
		{
			name: "Terraform with version constraint",
			fileContent: `terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      # depup package=aws-provider
      version = ">=4.0.0"
    }
  }
}
`,
			packages: []Package{{Name: "aws-provider", Version: "4.5.0"}},
			options:  FileUpdaterOptions{DryRun: false},
			expectedOutput: `terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      # depup package=aws-provider
      version = ">=4.5.0"
    }
  }
}
`,
			expectUpdated: true,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with test content
			tempFile, err := createTempFileWithContent(tt.fileContent, ".tf")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile)

			// Create updater
			updater := NewHclFileUpdater()

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

func TestHclFileUpdater_Supports(t *testing.T) {
	updater := NewHclFileUpdater()

	tests := []struct {
		extension string
		expected  bool
	}{
		{".hcl", true},
		{".tf", true},
		{".tfvars", true},
		{".yaml", false},
		{".yml", false},
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
