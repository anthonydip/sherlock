package parsers

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

type JestParser struct {
	filePath string
}

func NewJestParser(filePath string) *JestParser {
	return &JestParser{filePath: filePath}
}

func (j *JestParser) Parse() ([]TestFailure, error) {
	data, err := os.Open(j.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer data.Close()

	byteValue, err := io.ReadAll(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var output JestTestOutput
	if err := json.Unmarshal(byteValue, &output); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	var failures []TestFailure
	for _, suite := range output.TestResults {
		// First check suite-level message which might contain aggregated errors
		if len(suite.AssertionResults) == 0 && suite.Message != "" {
			if cleanMsg, location := extractErrorDetails(suite.Message); cleanMsg != "" {
				failures = append(failures, TestFailure{
					File:        suite.Name,
					TestName:    "Test Suite",
					Error:       cleanMsg,
					Location:    location,
					FullMessage: suite.Message,
				})
			}
		}

		// Then process individual test results
		for _, test := range suite.AssertionResults {
			if test.Status == "failed" {
				// Handle both FailureMessages and FailureDetails
				if len(test.FailureMessages) > 0 {
					for _, msg := range test.FailureMessages {
						failures = append(failures, extractFailure(suite, test, msg))
					}
				} else if len(test.FailureDetails) > 0 {
					for _, detail := range test.FailureDetails {
						if detail.MatcherResult != nil {
							msg := detail.MatcherResult.Message
							failures = append(failures, extractFailure(suite, test, msg))
						}
					}
				}
			}
		}
	}

	return failures, nil
}

func extractFailure(suite TestSuite, test AssertionResult, message string) TestFailure {
	cleanMsg := stripANSI(message)

	// First try to extract underlying error from matcher format
	if underlying := extractUnderlyingError(cleanMsg); underlying != "" {
		location := findLocation(cleanMsg)
		return TestFailure{
			File:        suite.Name,
			TestName:    buildTestName(test.AncestorTitles, test.Title),
			Error:       underlying,
			Location:    location,
			FullMessage: cleanMsg,
		}
	}

	// Fall back to standard error extraction
	errorMsg, location := extractErrorDetails(cleanMsg)
	return TestFailure{
		File:        suite.Name,
		TestName:    buildTestName(test.AncestorTitles, test.Title),
		Error:       errorMsg,
		Location:    location,
		FullMessage: cleanMsg,
	}
}

func stripANSI(message string) string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[mK]`)
	return ansiRegex.ReplaceAllString(message, "")
}

func extractUnderlyingError(message string) string {
	// Pattern 1: Error name/message format
	re1 := regexp.MustCompile(`Error name:\s+"([^"]+)"\s*\nError message:\s+"([^"]+)"`)
	if matches := re1.FindStringSubmatch(message); len(matches) > 2 {
		return fmt.Sprintf("%s: %s", matches[1], matches[2])
	}

	// Pattern 2: Error in assertion message
	re2 := regexp.MustCompile(`Expected.*?\n.*?([A-Za-z]+Error:[^\n]+)`)
	if matches := re2.FindStringSubmatch(message); len(matches) > 1 {
		return matches[1]
	}

	return ""
}

func extractErrorDetails(message string) (string, string) {
	// Clean ANSI codes first
	cleanMsg := stripANSI(message)

	// Get first line of error message
	lines := strings.Split(cleanMsg, "\n")
	if len(lines) == 0 {
		return "", ""
	}
	errorMsg := lines[0]

	// Try both stack trace formats
	location := findLocation(cleanMsg)

	return errorMsg, location
}

func findLocation(message string) string {
	// Format 1: "at function (file:line:column)"
	re1 := regexp.MustCompile(`at .*?\((.*?):(\d+):(\d+)\)`)
	// Format 2: "at file:line:column"
	re2 := regexp.MustCompile(`at (.*?):(\d+):(\d+)`)

	for _, line := range strings.Split(message, "\n") {
		// Try first format
		if matches := re1.FindStringSubmatch(line); len(matches) > 3 && isProjectFile(matches[1]) {
			return fmt.Sprintf("%s:%s", matches[1], matches[2])
		}
		// Try second format
		if matches := re2.FindStringSubmatch(line); len(matches) > 3 && isProjectFile(matches[1]) {
			return fmt.Sprintf("%s:%s", matches[1], matches[2])
		}
	}
	return ""
}

func isProjectFile(path string) bool {
	// Skip node_modules and test runner internals
	excluded := []string{"node_modules", "jest-circus", "jest-runner"}
	for _, e := range excluded {
		if strings.Contains(path, e) {
			return false
		}
	}
	return true
}

func extractLineNumber(line string) string {
	re := regexp.MustCompile(`:(\d+):\d+\)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return ":" + matches[1]
	}
	return ""
}

func buildTestName(ancestors []string, title string) string {
	if len(ancestors) > 0 {
		return strings.Join(ancestors, " > ") + " > " + title
	}
	return title
}

func (j *JestParser) RelevantFiles() []string {
	return []string{"*.js", "*.ts", "*.jsx", "*.tsx", "**/__tests__/*"}
}
