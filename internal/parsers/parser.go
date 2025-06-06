package parsers

import (
	"fmt"

	"github.com/anthonydip/sherlock/internal/git"
	"github.com/anthonydip/sherlock/internal/logger"
)

type TestFailure struct {
	File        string
	TestName    string
	Error       string
	Location    string
	FullMessage string
	LineNumber  int

	CodeChanges    string
	RelatedCommits []git.CommitInfo

	Context *TestFailureContext
}

type TestFailureContext struct {
	SurroundingCode string
	FullFileContent string
}

type Parser interface {
	Parse() ([]TestFailure, error)
	RelevantFiles() []string // Returns file patterns to check in git
}

func GetParser(name string, filePath string) (Parser, error) {
	logger.GlobalLogger.Debugf("Attempting to get parser for '%s' with name '%s'", filePath, name)
	switch name {
	case "jest":
		logger.GlobalLogger.Verbosef("Using Jest parser")
		return NewJestParser(filePath), nil
	case "auto":
		logger.GlobalLogger.Verbosef("Attempting parser auto-detection")
		return DetectParser(filePath)
	default:
		return nil, fmt.Errorf("unknown parser '%s'", name)
	}
}
