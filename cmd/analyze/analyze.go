package analyze

import (
	"fmt"

	"github.com/anthonydip/sherlock/internal/parsers"
	"github.com/spf13/cobra"
)

func NewAnalyzeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Diagnose test failures",
	}

	// Define analyze command flags
	cmd.Flags().StringP("test-output", "t", "", "Path to test results")
	cmd.Flags().StringP("api-key", "k", "", "OpenAI API key")
	cmd.Flags().StringP("parser", "p", "auto", "Test parser to use (jest, pytest, mocha, auto)")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		testOutput, _ := cmd.Flags().GetString("test-output")
		// apiKey, _ := cmd.Flags().GetString("api-key")
		parserName, _ := cmd.Flags().GetString("parser")

		// Select the correct parser
		parser, err := parsers.GetParser(parserName, testOutput)
		if err != nil {
			return fmt.Errorf("parser error: %w", err)
		}

		if parser == nil {
			fmt.Println("PARSER SELECTED!")
		}

		return nil
	}

	return cmd
}
