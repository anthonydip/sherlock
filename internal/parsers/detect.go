package parsers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func DetectParser(filePath string) (Parser, error) {
	// Get the absolute path based on the working directory
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("invalid path '%s': %v", filePath, err)
	}

	// Check if the file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		// Get the current working directory for error message
		wd, _ := os.Getwd()
		return nil, fmt.Errorf("file '%s' not found (looked in: %s)", filePath, wd)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	content := string(data)

	switch {
	case strings.Contains(content, "\"testResults\":"):
		return NewJestParser(filePath), nil
	default:
		return nil, fmt.Errorf("could not auto-detect parser")
	}
}
