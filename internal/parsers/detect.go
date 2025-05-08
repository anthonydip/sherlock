package parsers

import (
	"fmt"
	"os"
	"strings"
)

func DetectParser(filePath string) (Parser, error) {
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
