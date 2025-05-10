package analyze

import (
	"fmt"

	"github.com/anthonydip/sherlock/internal/parsers"
	"github.com/spf13/cobra"
)

func NewAnalyzeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze [test-output]",
		Short: "Diagnose test failures",
		Args:  cobra.ExactArgs(1),
	}

	// Define analyze command flags
	cmd.Flags().StringP("api-key", "k", "", "OpenAI API key")
	cmd.Flags().StringP("parser", "p", "auto", "Test parser to use (jest, pytest, mocha, auto)")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		testOutput := args[0]
		// apiKey, _ := cmd.Flags().GetString("api-key")
		parserName, _ := cmd.Flags().GetString("parser")

		// Select the correct parser
		parser, err := parsers.GetParser(parserName, testOutput)
		if err != nil {
			return fmt.Errorf("parser error: %w", err)
		}

		failures, err := parser.Parse()
		if err != nil {
			return fmt.Errorf("parser error: %w", err)
		}

		fmt.Println(failures)

		return nil
	}

	return cmd
}
