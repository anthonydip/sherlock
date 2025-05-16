package analyze

import (
	"fmt"
	"os"

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
				fmt.Fprintf(os.Stderr, "[ERROR] Requires exactly 1 test file\n\n")
				fmt.Fprintf(os.Stderr, "Usage: sherlock analyze [test-output] [flags]\n")
				return fmt.Errorf("Requires exactly 1 test file")
			}
			return nil
		},
	}

	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		fmt.Fprintf(os.Stderr, "[ERROR] Invalid flag: %v\n\n", err)
		fmt.Fprintf(os.Stderr, "Run '%s --help' for usage\n", cmd.CommandPath())
		return nil
	})

	// Define analyze command flags
	cmd.Flags().StringP("api-key", "k", "", "OpenAI API key")
	cmd.Flags().StringP("parser", "p", "auto", "Test parser to use (jest, pytest, mocha, auto)")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		testOutput := args[0]
		// apiKey, _ := cmd.Flags().GetString("api-key")
		parserName, _ := cmd.Flags().GetString("parser")

		logger.GlobalLogger.Debugf("Starting analysis of %s", testOutput)

		// Parser selection
		logger.GlobalLogger.Debugf("Selecting '%s' parser", parserName)
		parser, err := parsers.GetParser(parserName, testOutput)
		if err != nil {
			logger.GlobalLogger.Errorf("Parser selection failed: %v", err)
			return fmt.Errorf("test file not found: %s", testOutput)
		}

		// Parse test file
		failures, err := parser.Parse()
		if err != nil {
			logger.GlobalLogger.Errorf("Parsing failed: %v", err)
			return fmt.Errorf("parser error: %w", err)
		}

		if len(failures) > 0 {
			logger.GlobalLogger.Successf("Found %d test failures", len(failures))
		} else {
			logger.GlobalLogger.Successf("All test cases passed, no failures found")
			return nil
		}

		logger.GlobalLogger.Successf("Analysis completed")
		return nil
	}

	return cmd
}
