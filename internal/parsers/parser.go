package parsers

import "fmt"

type TestFailure struct {
	File        string
	TestName    string
	Error       string
	Location    string
	FullMessage string
}

type Parser interface {
	Parse() ([]TestFailure, error)
	RelevantFiles() []string // Returns file patterns to check in git
}

func GetParser(name string, filePath string) (Parser, error) {
	switch name {
	case "jest":
		return NewJestParser(filePath), nil
	case "auto":
		return DetectParser(filePath)
	default:
		return nil, fmt.Errorf("unknown parser: %s", name)
	}
}
