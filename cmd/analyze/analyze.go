package analyze

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/anthonydip/sherlock/internal/cli"
	"github.com/anthonydip/sherlock/internal/git"
	"github.com/anthonydip/sherlock/internal/logger"
	"github.com/anthonydip/sherlock/internal/parsers"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func NewAnalyzeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze [test-output]",
		Short: "Diagnose test failures",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				fmt.Fprintf(os.Stderr, "error: no test file specified\n")
				groups := generateOptionGroups(cmd)
				fmt.Fprint(os.Stderr, cli.FormatSubcommandUsage(cmd, groups))
				return fmt.Errorf("Requires exactly 1 test file")
			}
			return nil
		},
	}

	// Define analyze command flags
	cmd.Flags().StringP("api-key", "k", "", "OpenAI api key")
	cmd.Flags().StringP("parser", "p", "auto", "test parser to use (jest, pytest, mocha, auto)")
	cmd.Flags().Int("depth", 5, "maximum parent directory levels to search for .git (default = 5)")
	cmd.Flags().Bool("force", false, "proceed analysis with uncommitted changes")
	cmd.Flags().Bool("no-git", false, "skip Git integration entirely (repository detection and change analysis)")

	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		option := cli.StripInvalidFlag(err)

		groups := generateOptionGroups(cmd)

		fmt.Fprintf(os.Stderr, "unknown option: %s\n", option)
		fmt.Fprint(os.Stderr, cli.FormatSubcommandUsage(cmd, groups))
		return nil
	})

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		testOutput := args[0]
		// apiKey, _ := cmd.Flags().GetString("api-key")
		parserName, _ := cmd.Flags().GetString("parser")
		force, _ := cmd.Flags().GetBool("force")
		depth, _ := cmd.Flags().GetInt("depth")
		noGit, _ := cmd.Flags().GetBool("no-git")

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

		if !noGit {
			repo, err := git.OpenRepository(filepath.Dir(testOutput), depth)
			skipGit := err != nil
			if err != nil {
				if errors.Is(err, git.ErrNotAGitRepository) {
					logger.GlobalLogger.Verbosef("Unable to detect a Git repository within depth of %d (use --depth to change)", depth)
					logger.GlobalLogger.Warnf("Not running in a Git repository, skipping Git analysis")
				} else {
					logger.GlobalLogger.Errorf("Git error: %v", err)
					return fmt.Errorf("git error: %v", err)
				}
			}

			// Run Git analysis if running in a Git repository
			if !skipGit {
				// Check for uncommitted changes
				dirty, err := repo.IsDirty()
				if err != nil {
					logger.GlobalLogger.Errorf("Failed to check repo status: %v", err)
					return fmt.Errorf("git error: %v", err)
				}

				if dirty {
					if force {
						logger.GlobalLogger.Warnf("Uncommitted changes detected, proceeding with analysis")
					} else {
						logger.GlobalLogger.Errorf("Uncommitted changes detected (use --force to override)")
						return fmt.Errorf("uncommitted changes detected")
					}
				}
			}
		}

		logger.GlobalLogger.Successf("Analysis completed")
		return nil
	}

	return cmd
}

func generateOptionGroups(cmd *cobra.Command) []cli.FlagGroup {
	groups := []cli.FlagGroup{
		{
			Name: "Parser options",
			Flags: []*pflag.Flag{
				cmd.Flags().Lookup("parser"),
			},
		},
		{
			Name: "AI options",
			Flags: []*pflag.Flag{
				cmd.Flags().Lookup("api-key"),
			},
		},
	}

	return groups
}
