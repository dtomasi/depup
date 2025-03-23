package updater

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type YamlFileUpdater struct {
	supportedFileExtensions map[string]struct{}
	commentPattern          *regexp.Regexp
}

func NewYamlFileUpdater() *YamlFileUpdater {
	return &YamlFileUpdater{
		supportedFileExtensions: map[string]struct{}{
			".yaml": {},
			".yml":  {},
		},
		commentPattern: regexp.MustCompile(`#\s*depup\s+package=([^\s]+)`),
	}
}

func (u *YamlFileUpdater) Supports(fileExtension string) bool {
	_, ok := u.supportedFileExtensions[fileExtension]
	return ok
}

func (u *YamlFileUpdater) UpdateFile(filePath string, packages []Package, options FileUpdaterOptions) (string, bool, error) {
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
func (u *YamlFileUpdater) readFileContent(filePath string) ([]string, bool, error) {
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
func (u *YamlFileUpdater) processLines(lines []string, packages []Package, endsWithNewline bool) (string, bool) {
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
func (u *YamlFileUpdater) processInlineDepupComment(line string, packages []Package) (string, bool) {
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

	// Look for version in the line content
	versionMatches := VersionPattern.FindStringSubmatch(lineContent)
	if len(versionMatches) <= 3 {
		return line, false
	}

	// Try to update the version
	updatedContent, updated := u.updateVersion(lineContent, packageName, packages, versionMatches)
	if !updated {
		return line, false
	}

	// Reconstruct the line with updated version
	return updatedContent + comment, true
}

// processPreviousLineDepupComment handles the case where a depup comment is on the line before the version
func (u *YamlFileUpdater) processPreviousLineDepupComment(prevLine, currentLine string, packages []Package) (string, bool) {
	prevLineMatches := u.commentPattern.FindStringSubmatch(prevLine)
	if len(prevLineMatches) <= 1 {
		return currentLine, false
	}

	packageName := prevLineMatches[1]

	// Look for version in current line
	versionMatches := VersionPattern.FindStringSubmatch(currentLine)
	if len(versionMatches) <= 3 {
		return currentLine, false
	}

	// Try to update the version
	updatedContent, updated := u.updateVersion(currentLine, packageName, packages, versionMatches)
	if !updated {
		return currentLine, false
	}

	return updatedContent, true
}

// updateVersion updates the version in a line if the package name matches
func (u *YamlFileUpdater) updateVersion(line, packageName string, packages []Package, versionMatches []string) (string, bool) {
	for _, pkg := range packages {
		if pkg.Name == packageName {
			startQuote := versionMatches[1]
			currentVersion := versionMatches[2]
			endQuote := versionMatches[3]

			// If the versions are the same (ignoring constraints), no update needed
			cleanVersion := regexp.MustCompile(`[>=~<^]+\s*`).ReplaceAllString(currentVersion, "")
			if cleanVersion == pkg.Version {
				return line, false
			}

			// Replace with new version, removing any constraints
			updatedLine := strings.Replace(
				line,
				versionMatches[0],
				startQuote+pkg.Version+endQuote,
				1,
			)

			return updatedLine, true
		}
	}

	return line, false
}
