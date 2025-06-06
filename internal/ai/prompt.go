package ai

import (
	"fmt"
	"strings"

	"github.com/anthonydip/sherlock/internal/parsers"
)

func GeneratePrompt(failure parsers.TestFailure) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Test Failure Analysis Request\n\n"))
	sb.WriteString(fmt.Sprintf("Test Name: %s\n", failure.TestName))
	sb.WriteString(fmt.Sprintf("Error Message: %s\n", failure.Error))

	if failure.FullMessage != "" {
		sb.WriteString(fmt.Sprintf("Full Error Message: %s\n", failure.FullMessage))
	}

	sb.WriteString(fmt.Sprintf("\nFile: %s (Line %d)\n", failure.Location, failure.LineNumber))

	if failure.Context.SurroundingCode != "" {
		sb.WriteString(fmt.Sprintf("\nCode Context:\n%s\n", failure.Context.SurroundingCode))
	}

	if failure.CodeChanges != "" {
		sb.WriteString(fmt.Sprintf("\nRecent Line Changes:\n%s\n", failure.CodeChanges))
	}

	sb.WriteString("\nPlease analyze this test failure and:\n")
	sb.WriteString("1. Explain the likely root cause\n")
	sb.WriteString("2. Suggest specific fixes if possible\n")
	sb.WriteString("3. Note any relevant patterns from the recent changes\n")
	sb.WriteString("4. Provide code examples if applicable\n")

	return sb.String()
}
