package parsers

import "fmt"

type TestFailure struct {
	Name    string
	Message string
	File    string
	Line    int
}

type Parser interface {
	Parse() ([]TestFailure, error)
	RelevantFiles() []string // Returns file patterns to check in git
}

func GetParser(name string, filePath string) (Parser, error) {
	switch name {
	case "jest":
		fmt.Println("JEST PARSER SELECTED")
		return NewJestParser(filePath), nil
	case "auto":
		return DetectParser(filePath)
	default:
		return nil, fmt.Errorf("unknown parser: %s", name)
	}
}
