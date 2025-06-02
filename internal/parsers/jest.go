package parsers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/anthonydip/sherlock/internal/logger"
)

type JestParser struct {
	filePath string
}

func NewJestParser(filePath string) *JestParser {
	return &JestParser{
		filePath: filePath,
	}
}

func (j *JestParser) Parse() ([]TestFailure, error) {
	logger.GlobalLogger.Debugf("Parsing Jest output from: %s", j.filePath)

	data, err := os.Open(j.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("test file does not exist: %q", j.filePath)
		} else if os.IsPermission(err) {
			return nil, fmt.Errorf("no permission to read file: %q", j.filePath)
		}
		return nil, fmt.Errorf("failed to access test file: %w", err)
	}
	defer data.Close()

	byteValue, err := io.ReadAll(data)
	if err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, fmt.Errorf("file appears truncated or corrupted")
		}
		if pathErr := (&os.PathError{}); errors.As(err, &pathErr) {
			return nil, fmt.Errorf("lost access to file while reading: %w", err)
		}
		return nil, fmt.Errorf("failed to read test file contents: %w", err)
	}
	logger.GlobalLogger.Debugf("Read %d bytes from test file", len(byteValue))

	var output JestTestOutput
	if err := json.Unmarshal(byteValue, &output); err != nil {
		return nil, fmt.Errorf("expected valid JSON test output, but got malformed data")
	}

	logger.GlobalLogger.Verbosef("Found %d test suite(s)", len(output.TestResults))

	var failures []TestFailure
	for _, suite := range output.TestResults {
		// NOTE: suite.Name gives the name of the file with the test cases
		logger.GlobalLogger.Debugf("Processing suite: %s", suite.Name)

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
				logger.GlobalLogger.Verbosef("Processing failed test: %s", test.Title)

				// Handle both FailureMessages and FailureDetails
				if len(test.FailureMessages) > 0 {
					for _, msg := range test.FailureMessages {
						extractedFail := extractFailure(suite, test, msg)
						logger.GlobalLogger.Debugf("Extracted file path: %s", extractedFail.File)
						logger.GlobalLogger.Debugf("Extracted test name: %s", extractedFail.TestName)
						logger.GlobalLogger.Debugf("Extracted error: %s", extractedFail.Error)
						logger.GlobalLogger.Debugf("Extracted location: %s", extractedFail.Location)
						logger.GlobalLogger.Debugf("Extracted line number: %d", extractedFail.LineNumber)
						failures = append(failures, extractedFail)
					}
				} else if len(test.FailureDetails) > 0 {
					for _, detail := range test.FailureDetails {
						if detail.MatcherResult != nil {
							msg := detail.MatcherResult.Message
							extractedFail := extractFailure(suite, test, msg)
							logger.GlobalLogger.Debugf("Extracted file path: %s", extractedFail.File)
							logger.GlobalLogger.Debugf("Extracted test name: %s", extractedFail.TestName)
							logger.GlobalLogger.Debugf("Extracted error: %s", extractedFail.Error)
							logger.GlobalLogger.Debugf("Extracted location: %s", extractedFail.Location)
							logger.GlobalLogger.Debugf("Extracted line number: %d", extractedFail.LineNumber)
							failures = append(failures, extractedFail)
						}
					}
				}
			}
		}
	}

	logger.GlobalLogger.Verbosef("Found %d total failures", len(failures))
	return failures, nil
}

func extractFailure(suite TestSuite, test AssertionResult, message string) TestFailure {
	cleanMsg := stripANSI(message)

	// First try to extract underlying error from matcher format
	if underlying := extractUnderlyingError(cleanMsg); underlying != "" {
		location := findLocation(cleanMsg)
		lineNumber := extractLineNumber(location)
		return TestFailure{
			File:        suite.Name,
			TestName:    buildTestName(test.AncestorTitles, test.Title),
			Error:       underlying,
			Location:    location,
			FullMessage: cleanMsg,
			LineNumber:  lineNumber,
		}
	}

	// Fall back to standard error extraction
	errorMsg, location := extractErrorDetails(cleanMsg)
	lineNumber := extractLineNumber(location)
	return TestFailure{
		File:        suite.Name,
		TestName:    buildTestName(test.AncestorTitles, test.Title),
		Error:       errorMsg,
		Location:    location,
		FullMessage: cleanMsg,
		LineNumber:  lineNumber,
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

func normalizePath(path string) string {
	// If path is already absolute (starts with drive letter or slash)
	if filepath.IsAbs(path) {
		return path
	}

	// If path is relative but contains standard Jest paths
	if strings.HasPrefix(path, "src/") || strings.HasPrefix(path, "tests/") {
		// Make it absolute by joining with current working directory
		if absPath, err := filepath.Abs(path); err == nil {
			return absPath
		}
	}

	// Default return (will handle cases like "src/file.js:123")
	return path
}

func findLocation(message string) string {
	// Format 1: "at function (file:line:column)"
	re1 := regexp.MustCompile(`at .*?\((.*?):(\d+):(\d+)\)`)
	// Format 2: "at file:line:column"
	re2 := regexp.MustCompile(`at (.*?):(\d+):(\d+)`)

	for _, line := range strings.Split(message, "\n") {
		// Try first format
		if matches := re1.FindStringSubmatch(line); len(matches) > 3 && isProjectFile(matches[1]) {
			return normalizePath(fmt.Sprintf("%s:%s", matches[1], matches[2]))
		}
		// Try second format
		if matches := re2.FindStringSubmatch(line); len(matches) > 3 && isProjectFile(matches[1]) {
			return normalizePath(fmt.Sprintf("%s:%s", matches[1], matches[2]))
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

func extractLineNumber(location string) int {
	result := strings.Split(location, ":")
	i, err := strconv.Atoi(result[len(result)-1])
	if err != nil {
		logger.GlobalLogger.Debugf("Failed to extract line number from: %s", location)
		return 0
	}

	return i
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
