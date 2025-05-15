package analyze

import (
	"fmt"

	"github.com/anthonydip/sherlock/internal/logger"
	"github.com/anthonydip/sherlock/internal/parsers"
	"github.com/spf13/cobra"
)

func NewAnalyzeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze [test-output]",
		Short: "Diagnose test failures",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("Analyze requires exactly 1 argument (test output)\n\nUsage: %s", cmd.UsageString())
			}
			return nil
		},
	}

	// Define analyze command flags
	cmd.Flags().StringP("api-key", "k", "", "OpenAI API key")
	cmd.Flags().StringP("parser", "p", "auto", "Test parser to use (jest, pytest, mocha, auto)")
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		testOutput := args[0]
		// apiKey, _ := cmd.Flags().GetString("api-key")
		parserName, _ := cmd.Flags().GetString("parser")
		verbose, _ := cmd.Flags().GetBool("verbose")

		logger.GlobalLogger.SetVerbose(verbose)

		// Select the correct parser
		parser, err := parsers.GetParser(parserName, testOutput)
		if err != nil {
			return fmt.Errorf("use 'sherlock analyze --help' for more information")
		}

		failures, err := parser.Parse()
		if err != nil {
			return fmt.Errorf("parser error: %w", err)
		}

		for _, failure := range failures {
			fmt.Printf("File: %s\n", failure.File)
			fmt.Printf("TestName: %s\n", failure.TestName)
			fmt.Printf("Error: %s\n", failure.Error)
			fmt.Printf("Location: %s\n", failure.Location)
			fmt.Println("-----")
		}

		return nil
	}

	return cmd
}
