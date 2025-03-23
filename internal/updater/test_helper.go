package updater

import "os"

func createTempFileWithContent(content string, ext string) (string, error) {
	tempFile, err := os.CreateTemp("", ext+"updater_test_*"+ext)
	if err != nil {
		return "", err
	}

	if _, err := tempFile.WriteString(content); err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}

	if err := tempFile.Close(); err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}

	return tempFile.Name(), nil
}
