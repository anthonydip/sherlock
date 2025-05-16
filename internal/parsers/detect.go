package parsers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthonydip/sherlock/internal/logger"
)

func DetectParser(filePath string) (Parser, error) {
	logger.GlobalLogger.Debugf("Attempting to detect parser for file: %s", filePath)

	// Get the absolute path based on the working directory
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("invalid path '%s': %v", filePath, err)
	}

	logger.GlobalLogger.Verbosef("Resolved absolute path: %s", absPath)

	// Check if the file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		// Get the current working directory for error message
		wd, _ := os.Getwd()
		return nil, fmt.Errorf("file '%s' not found (looked in: %s)", filePath, wd)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		logger.GlobalLogger.Errorf("File read error: %v", err)
		return nil, err
	}

	content := string(data)
	logger.GlobalLogger.Debugf("Read %d bytes from file", len(content))

	switch {
	case strings.Contains(content, "\"testResults\":"):
		logger.GlobalLogger.Verbosef("Detected Jest test format")
		return NewJestParser(filePath), nil
	default:
		logger.GlobalLogger.Errorf("Auto-detection failed")
		return nil, fmt.Errorf("could not auto-detect parser")
	}
}
