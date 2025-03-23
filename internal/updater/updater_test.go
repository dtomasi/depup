package updater

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// MockFileUpdater is a mock implementation of the FileUpdater interface for testing
type MockFileUpdater struct {
	supportedExtensions []string
	updatedFiles        map[string]string
	shouldFail          bool
	shouldUpdate        bool
}

func NewMockFileUpdater(extensions []string, shouldFail, shouldUpdate bool) *MockFileUpdater {
	return &MockFileUpdater{
		supportedExtensions: extensions,
		updatedFiles:        make(map[string]string),
		shouldFail:          shouldFail,
		shouldUpdate:        shouldUpdate,
	}
}

func (m *MockFileUpdater) GetSupportedExtensions() []string {
	return m.supportedExtensions
}

func (m *MockFileUpdater) Supports(fileExtension string) bool {
	for _, ext := range m.supportedExtensions {
		if ext == fileExtension {
			return true
		}
	}
	return false
}

func (m *MockFileUpdater) UpdateFile(filePath string, packages []Package, options FileUpdaterOptions) (string, bool, error) {
	if m.shouldFail {
		return "", false, errors.New("mock update failure")
	}

	updatedContent := "updated content"
	m.updatedFiles[filePath] = updatedContent
	return updatedContent, m.shouldUpdate, nil
}

func TestNewUpdater(t *testing.T) {
	tests := []struct {
		name     string
		options  []Option
		expected *Updater
	}{
		{
			name:    "default options",
			options: []Option{},
			expected: &Updater{
				dryRun:         false,
				recursive:      true,
				fileExtensions: []string{},
			},
		},
		{
			name: "with dry run",
			options: []Option{
				WithDryRun(true),
			},
			expected: &Updater{
				dryRun:         true,
				recursive:      true,
				fileExtensions: []string{".yaml", ".yml"},
			},
		},
		{
			name: "with non-recursive",
			options: []Option{
				WithRecursive(false),
			},
			expected: &Updater{
				dryRun:         false,
				recursive:      false,
				fileExtensions: []string{".yaml", ".yml"},
			},
		},
		{
			name: "with custom extensions",
			options: []Option{
				WithFileExtensions([]string{".json", ".conf"}),
			},
			expected: &Updater{
				dryRun:         false,
				recursive:      true,
				fileExtensions: []string{".json", ".conf"},
			},
		},
		{
			name: "with multiple options",
			options: []Option{
				WithDryRun(true),
				WithRecursive(false),
				WithFileExtensions([]string{".toml"}),
			},
			expected: &Updater{
				dryRun:         true,
				recursive:      false,
				fileExtensions: []string{".toml"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updater := NewUpdater(tt.options...)

			// Check updater has one default updater
			if len(updater.updaters) != 2 {
				t.Errorf("expected 2 default updater, got %d", len(updater.updaters))
			}

			// Check configuration
			if updater.dryRun != tt.expected.dryRun {
				t.Errorf("dryRun: expected %v, got %v", tt.expected.dryRun, updater.dryRun)
			}

			if updater.recursive != tt.expected.recursive {
				t.Errorf("recursive: expected %v, got %v", tt.expected.recursive, updater.recursive)
			}
		})
	}
}

func TestUpdater_isFileExtensionSupported(t *testing.T) {
	updater := NewUpdater(WithFileExtensions([]string{".yaml", ".yml", ".json"}))

	tests := []struct {
		path     string
		expected bool
	}{
		{"test.yaml", true},
		{"test.yml", true},
		{"test.json", true},
		{"test.toml", false},
		{"test.txt", false},
		{"test", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := updater.isFileExtensionSupported(tt.path)
			if result != tt.expected {
				t.Errorf("isFileExtensionSupported(%q): expected %v, got %v", tt.path, tt.expected, result)
			}
		})
	}
}

func TestUpdater_getFileUpdater(t *testing.T) {
	mockYaml := NewMockFileUpdater([]string{".yaml", ".yml"}, false, true)
	mockJson := NewMockFileUpdater([]string{".json"}, false, true)

	updater := NewUpdater()
	updater.updaters = []FileUpdater{mockYaml, mockJson}

	tests := []struct {
		extension string
		wantErr   bool
	}{
		{".yaml", false},
		{".yml", false},
		{".json", false},
		{".toml", true},
		{".txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.extension, func(t *testing.T) {
			updater, err := updater.getFileUpdater(tt.extension)

			if tt.wantErr {
				if err == nil {
					t.Errorf("getFileUpdater(%q): expected error, got nil", tt.extension)
				}
				return
			}

			if err != nil {
				t.Errorf("getFileUpdater(%q): unexpected error: %v", tt.extension, err)
				return
			}

			if updater == nil {
				t.Errorf("getFileUpdater(%q): got nil updater", tt.extension)
			}

			if !updater.Supports(tt.extension) {
				t.Errorf("getFileUpdater(%q): returned updater that doesn't support the extension", tt.extension)
			}
		})
	}
}

func TestUpdater_Update_SingleFile(t *testing.T) {
	// Create a temporary test file
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.yaml")
	err := os.WriteFile(filePath, []byte("original content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	mockUpdater := NewMockFileUpdater([]string{".yaml", ".yml"}, false, true)

	updater := NewUpdater()
	updater.updaters = []FileUpdater{mockUpdater}

	packages := []Package{
		{Name: "example", Version: "1.0.0"},
	}

	// Test successful update
	err = updater.Update(filePath, packages)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if _, ok := mockUpdater.updatedFiles[filePath]; !ok {
		t.Errorf("expected file %s to be updated", filePath)
	}

	// Test with failing updater
	mockUpdater = NewMockFileUpdater([]string{".yaml", ".yml"}, true, true)
	updater.updaters = []FileUpdater{mockUpdater}

	err = updater.Update(filePath, packages)
	if err == nil {
		t.Errorf("expected Update to fail with failing updater")
	}
}

func TestUpdater_Update_Directory(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create files in root dir
	err := os.WriteFile(filepath.Join(tempDir, "root1.yaml"), []byte("content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "root2.yml"), []byte("content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "root3.txt"), []byte("content"), 0644) // Not supported
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create subdirectory with files
	subDir := filepath.Join(tempDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	err = os.WriteFile(filepath.Join(subDir, "sub1.yaml"), []byte("content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	err = os.WriteFile(filepath.Join(subDir, "sub2.yml"), []byte("content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	mockUpdater := NewMockFileUpdater([]string{".yaml", ".yml"}, false, true)

	// Test with recursive mode (default)
	updater := NewUpdater()
	updater.updaters = []FileUpdater{mockUpdater}

	packages := []Package{{Name: "example", Version: "1.0.0"}}

	err = updater.Update(tempDir, packages)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Should have updated 4 files (2 in root, 2 in subdir)
	if len(mockUpdater.updatedFiles) != 4 {
		t.Errorf("expected 4 files to be updated, got %d", len(mockUpdater.updatedFiles))
	}

	// Test with non-recursive mode
	mockUpdater = NewMockFileUpdater([]string{".yaml", ".yml"}, false, true)
	updater = NewUpdater(WithRecursive(false))
	updater.updaters = []FileUpdater{mockUpdater}

	err = updater.Update(tempDir, packages)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Should have updated only 2 files in root dir
	if len(mockUpdater.updatedFiles) != 2 {
		t.Errorf("expected 2 files to be updated in non-recursive mode, got %d", len(mockUpdater.updatedFiles))
	}

	// Verify sub-directory files were not processed
	if _, ok := mockUpdater.updatedFiles[filepath.Join(subDir, "sub1.yaml")]; ok {
		t.Errorf("expected file in subdirectory to not be updated in non-recursive mode")
	}
}

func TestUpdater_Update_DryRun(t *testing.T) {
	// Create a temporary test file
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.yaml")
	originalContent := "original content"
	err := os.WriteFile(filePath, []byte(originalContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	mockUpdater := NewMockFileUpdater([]string{".yaml", ".yml"}, false, true)
	updater := NewUpdater(WithDryRun(true))
	updater.updaters = []FileUpdater{mockUpdater}

	packages := []Package{{Name: "example", Version: "1.0.0"}}

	err = updater.Update(filePath, packages)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// File should be in updated map but actual file content shouldn't change
	if _, ok := mockUpdater.updatedFiles[filePath]; !ok {
		t.Errorf("expected file to be processed in dry run mode")
	}

	// Read the file back to ensure it wasn't modified
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	if string(content) != originalContent {
		t.Errorf("file was modified in dry run mode")
	}
}

func TestUpdater_Update_Errors(t *testing.T) {
	tempDir := t.TempDir()

	updater := NewUpdater()
	mockUpdater := NewMockFileUpdater([]string{".yaml", ".yml"}, false, true)
	updater.updaters = []FileUpdater{mockUpdater}

	packages := []Package{{Name: "example", Version: "1.0.0"}}

	// Test with non-existent path
	nonExistentPath := filepath.Join(tempDir, "doesnotexist")
	err := updater.Update(nonExistentPath, packages)
	if err == nil {
		t.Errorf("expected error for non-existent path")
	}

	// Test with recursive flag on file (should error)
	filePath := filepath.Join(tempDir, "test.yaml")
	err = os.WriteFile(filePath, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	updater = NewUpdater(WithRecursive(true))
	updater.updaters = []FileUpdater{mockUpdater}

	// This shouldn't error as the implementation logic allows recursive flag on files
	err = updater.Update(filePath, packages)
	if err != nil {
		t.Errorf("unexpected error when using recursive flag with file: %v", err)
	}

	// Test with unsupported file extension
	unsupportedPath := filepath.Join(tempDir, "unsupported.txt")
	err = os.WriteFile(unsupportedPath, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	err = updater.Update(unsupportedPath, packages)
	if err != nil {
		t.Errorf("expected silent skip for unsupported file extension, got error: %v", err)
	}

	// The unsupported file shouldn't be in the updated map
	if _, ok := mockUpdater.updatedFiles[unsupportedPath]; ok {
		t.Errorf("unsupported file was updated")
	}
}
