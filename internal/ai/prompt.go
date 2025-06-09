package ai

import (
	"fmt"
	"strings"

	"github.com/anthonydip/sherlock/internal/parsers"
)

func GeneratePrompt(failure parsers.TestFailure) string {
	var sb strings.Builder

	sb.WriteString("As a senior engineer, analyze this test failure and respond EXACTLY in this format:\n\n")
	sb.WriteString("### Root Cause\n[1-3 sentence explanation]\n\n")
	sb.WriteString("### Suggested Fixes\n- [Fix 1]\n- [Fix 2]\n\n")
	sb.WriteString("### Code Example\n```[language]\n[Relevant code snippet]\n```\n\n")
	sb.WriteString("---\n")

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

	return sb.String()
}

func GenerateBatchPrompt(failures []parsers.TestFailure) string {
	var sb strings.Builder

	sb.WriteString("Analyze these test failures concisely. For EACH failure, respond in this EXACT format:\n\n")
	sb.WriteString("#### [Test Name]\n")
	sb.WriteString("**Root Cause**: [1 sentence]\n")
	sb.WriteString("**Quick Fix**: [1-2 bullet points]\n")
	sb.WriteString("**Relevant Code**: \n```[language]\n[Key lines only]\n```\n\n")
	sb.WriteString("---\n")

	for i, failure := range failures {
		sb.WriteString(fmt.Sprintf("### Failure %d/%d\n", i+1, len(failures)))
		sb.WriteString(fmt.Sprintf("Test: %s\n", failure.TestName))
		sb.WriteString(fmt.Sprintf("Error: %s\n", failure.Error))
		sb.WriteString(fmt.Sprintf("Location: %s:%d\n", failure.Location, failure.LineNumber))

		if failure.Context.SurroundingCode != "" {
			sb.WriteString(fmt.Sprintf("Code Context:\n%s\n", failure.Context.SurroundingCode))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
