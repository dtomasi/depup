package updater

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type DotEnvFileUpdater struct {
	supportedFileExtensions map[string]struct{}
	commentPattern          *regexp.Regexp
}

func NewDotEnvFileUpdater() *DotEnvFileUpdater {
	return &DotEnvFileUpdater{
		supportedFileExtensions: map[string]struct{}{
			".env":   {},
			".env.*": {},
		},
		commentPattern: regexp.MustCompile(`#\s*depup\s+package=([^\s]+)`),
	}
}

func (u *DotEnvFileUpdater) Supports(fileExtension string) bool {
	// Check for direct extension match first
	if _, ok := u.supportedFileExtensions[fileExtension]; ok {
		return true
	}

	// Check for glob pattern matches
	fileName := "file" + fileExtension // Create a fake filename to match against patterns
	for pattern := range u.supportedFileExtensions {
		if strings.Contains(pattern, "*") || strings.Contains(pattern, "?") {
			// Handle glob patterns
			matched, err := filepath.Match(pattern, fileExtension)
			if err == nil && matched {
				return true
			}

			// Try matching against the full filename for extensions like .env.*
			matched, err = filepath.Match(pattern, fileName)
			if err == nil && matched {
				return true
			}
		}
	}
	return false
}

func (u *DotEnvFileUpdater) GetSupportedExtensions() []string {
	extensions := make([]string, 0, len(u.supportedFileExtensions))
	for ext := range u.supportedFileExtensions {
		extensions = append(extensions, ext)
	}
	return extensions
}

func (u *DotEnvFileUpdater) UpdateFile(filePath string, packages []Package, options FileUpdaterOptions) (string, bool, error) {
	// Read file and prepare data
	lines, endsWithNewline, err := u.readFileContent(filePath)
	if err != nil {
		return "", false, err
	}

	// Process lines and build output
	outputContent, updated := u.processLines(lines, packages, endsWithNewline)

	// Write changes if needed
	if updated && !options.DryRun {
		err = os.WriteFile(filePath, []byte(outputContent), 0644)
		if err != nil {
			return "", false, fmt.Errorf("failed to write updated content to %s: %w", filePath, err)
		}
	}

	return outputContent, updated, nil
}

// readFileContent reads a file and returns its lines and whether it ends with a newline
func (u *DotEnvFileUpdater) readFileContent(filePath string) ([]string, bool, error) {
	// Read file content first to check for trailing newline
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, false, fmt.Errorf("cannot read file %s: %w", filePath, err)
	}

	// Check if file ends with newline
	endsWithNewline := len(fileContent) > 0 && (fileContent[len(fileContent)-1] == '\n')

	// Open file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return nil, false, fmt.Errorf("cannot open file %s: %w", filePath, err)
	}
	defer file.Close()

	// Create a scanner to read line by line
	scanner := bufio.NewScanner(file)

	// Read lines
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Check for scanner errors
	if err = scanner.Err(); err != nil {
		return nil, false, fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	return lines, endsWithNewline, nil
}

// processLines processes all lines and returns the modified content and update status
func (u *DotEnvFileUpdater) processLines(lines []string, packages []Package, endsWithNewline bool) (string, bool) {
	var output strings.Builder
	updated := false

	for i := 0; i < len(lines); i++ {
		currentLine := lines[i]
		modifiedLine := currentLine
		lineUpdated := false

		// Check for inline depup comment
		if newLine, lineWasUpdated := u.processInlineDepupComment(currentLine, packages); lineWasUpdated {
			modifiedLine = newLine
			lineUpdated = true
		} else if i > 0 {
			// Check for depup comment in previous line
			if newLine, lineWasUpdated := u.processPreviousLineDepupComment(lines[i-1], currentLine, packages); lineWasUpdated {
				modifiedLine = newLine
				lineUpdated = true
			}
		}

		// Add the current line to output
		output.WriteString(modifiedLine)

		// Add newline if not the last line or if the original file had a trailing newline
		if i < len(lines)-1 || endsWithNewline {
			output.WriteString("\n")
		}

		// Update the overall update status
		updated = updated || lineUpdated
	}

	return output.String(), updated
}

// processInlineDepupComment handles the case where a depup comment is on the same line as the version
func (u *DotEnvFileUpdater) processInlineDepupComment(line string, packages []Package) (string, bool) {
	// Don't process lines that are only comments
	if strings.TrimSpace(line) == "" || strings.TrimSpace(line)[0] == '#' {
		return line, false
	}

	inlineCommentRegex := regexp.MustCompile(`(.*?)(\s*#.*)$`)
	inlineMatches := inlineCommentRegex.FindStringSubmatch(line)

	if len(inlineMatches) <= 2 {
		return line, false
	}

	lineContent := inlineMatches[1]
	comment := inlineMatches[2]

	// Check if it's a depup comment
	depupMatches := u.commentPattern.FindStringSubmatch(comment)
	if len(depupMatches) <= 1 {
		return line, false
	}

	// This is a depup comment
	packageName := depupMatches[1]

	// Parse KEY=VALUE format preserving spaces
	keyValueRegex := regexp.MustCompile(`^([^=]+)(=)(.*)$`)
	keyValueMatches := keyValueRegex.FindStringSubmatch(lineContent)
	if len(keyValueMatches) <= 3 {
		return line, false
	}

	key := keyValueMatches[1]
	equals := keyValueMatches[2]
	value := keyValueMatches[3]

	// Try to update the version
	updatedValue, updated := u.updateEnvValue(value, packageName, packages)
	if !updated {
		return line, false
	}

	// Reconstruct the line with updated version
	return key + equals + updatedValue + comment, true
}

// processPreviousLineDepupComment handles the case where a depup comment is on the line before the version
func (u *DotEnvFileUpdater) processPreviousLineDepupComment(prevLine, currentLine string, packages []Package) (string, bool) {
	// Skip if previous line is not a depup comment or current line is a comment
	if strings.TrimSpace(currentLine) == "" || strings.TrimSpace(currentLine)[0] == '#' {
		return currentLine, false
	}

	prevLineMatches := u.commentPattern.FindStringSubmatch(prevLine)
	if len(prevLineMatches) <= 1 {
		return currentLine, false
	}

	packageName := prevLineMatches[1]

	// Parse KEY=VALUE format preserving spaces
	keyValueRegex := regexp.MustCompile(`^([^=]+)(=)(.*)$`)
	keyValueMatches := keyValueRegex.FindStringSubmatch(currentLine)
	if len(keyValueMatches) <= 3 {
		return currentLine, false
	}

	key := keyValueMatches[1]
	equals := keyValueMatches[2]
	value := keyValueMatches[3]

	// Try to update the version
	updatedValue, updated := u.updateEnvValue(value, packageName, packages)
	if !updated {
		return currentLine, false
	}

	return key + equals + updatedValue, true
}

// updateEnvValue updates the version value if the package name matches
func (u *DotEnvFileUpdater) updateEnvValue(value, packageName string, packages []Package) (string, bool) {
	for _, pkg := range packages {
		if pkg.Name == packageName {
			// Handle quoted values
			quotedValueRegex := regexp.MustCompile(`^(\s*)(['"])(.*?)(['"])(.*)$`)
			quotedMatches := quotedValueRegex.FindStringSubmatch(value)

			if len(quotedMatches) > 4 {
				// Value is quoted
				leadingSpace := quotedMatches[1]
				startQuote := quotedMatches[2]
				currentValue := quotedMatches[3]
				endQuote := quotedMatches[4]
				trailingContent := quotedMatches[5]

				// Only update if it matches the version pattern
				if !versionPattern.MatchString(currentValue) {
					return value, false
				}

				if currentValue == pkg.Version {
					return value, false
				}

				return leadingSpace + startQuote + pkg.Version + endQuote + trailingContent, true
			} else {
				// Value is not quoted - extract just the version part
				spaceAndVersionRegex := regexp.MustCompile(`^(\s*)([^\s]+)(.*)$`)
				spaceMatches := spaceAndVersionRegex.FindStringSubmatch(value)

				if len(spaceMatches) > 3 {
					leadingSpace := spaceMatches[1]
					currentValue := spaceMatches[2]
					trailingContent := spaceMatches[3]

					// Only update if it matches the version pattern
					if !versionPattern.MatchString(currentValue) {
						return value, false
					}

					if currentValue == pkg.Version {
						return value, false
					}

					return leadingSpace + pkg.Version + trailingContent, true
				}
			}
		}
	}

	return value, false
}
