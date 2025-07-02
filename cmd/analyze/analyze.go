package analyze

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthonydip/sherlock/internal/ai"
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

	// Parser flags
	cmd.Flags().StringP("parser", "p", "auto", "test parser to use (jest, pytest, mocha, auto)")

	// Git flags
	cmd.Flags().Int("git-depth", 5, "maximum parent directory levels to search for .git (default: 5)")
	cmd.Flags().Int("context-lines", 3, "number of surrounding code lines to include in analysis (default: 3)")
	cmd.Flags().Int("commit-depth", 3, "number of historical commits to analyze (default: 3)")
	cmd.Flags().Bool("force", false, "proceed analysis with uncommitted changes")
	cmd.Flags().Bool("no-git", false, "skip Git integration entirely (repository detection and change analysis)")

	// AI flags
	cmd.Flags().StringP("api-key", "k", "", "AI API key override (openai|groq)")
	cmd.Flags().StringP("model", "m", "", "ai model to use (default: gpt-3.5-turbo|llama3-70b-8192)")
	cmd.Flags().String("ai-provider", "", "ai provider to use (openai|groq)")
	cmd.Flags().BoolP("batch", "b", false, "batch multiple test failures into one AI request (default: false)")

	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		option := cli.StripInvalidFlag(err)

		groups := generateOptionGroups(cmd)

		fmt.Fprintf(os.Stderr, "unknown option: %s\n", option)
		fmt.Fprint(os.Stderr, cli.FormatSubcommandUsage(cmd, groups))
		return nil
	})

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		testOutput := args[0]
		parserName, _ := cmd.Flags().GetString("parser")
		force, _ := cmd.Flags().GetBool("force")
		depth, _ := cmd.Flags().GetInt("git-depth")
		contextLines, _ := cmd.Flags().GetInt("context-lines")
		commitDepth, _ := cmd.Flags().GetInt("commit-depth")
		noGit, _ := cmd.Flags().GetBool("no-git")
		batch, _ := cmd.Flags().GetBool("batch")

		aiOpts, err := getAIOptions(cmd)
		if err != nil {
			logger.GlobalLogger.Errorf("%s", err)
			return err
		}

		logger.GlobalLogger.Debugf("Starting analysis of %s", testOutput)

		logger.GlobalLogger.Debugf("Detected AI options: %s and %s", aiOpts.Provider, aiOpts.Model)

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

		if noGit {
			logger.GlobalLogger.Verbosef("--no-git used, skipping Git integration")
		} else {

			// Attempt to open the Git repository
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

				// If any uncommitted changes were found
				if dirty {
					if force {
						logger.GlobalLogger.Warnf("Uncommitted changes detected, proceeding with analysis")
					} else {
						logger.GlobalLogger.Errorf("Uncommitted changes detected (use --force to override)")
						return fmt.Errorf("uncommitted changes detected")
					}
				}

				// Get commit history for the affected files
				for index := range failures {
					failure := &failures[index]

					// Convert absolute path to repo-relative path
					relPath, err := git.NormalizeTestPath(failure.Location, repo.Path())
					if err != nil {
						logger.GlobalLogger.Errorf("Failure %d - Failed to normalize path: %v", index+1, err)
						continue
					}

					logger.GlobalLogger.Debugf("Failure %d - Analyzing failure in: %s", index+1, relPath)

					// Get Git commit history for the affected file
					commitHistory, err := repo.GetEnhancedFileHistory(relPath, commitDepth)
					if err != nil {
						logger.GlobalLogger.Errorf("Failure %d - Failed to get commit history: %v", index+1, err)
						return err
					}

					// Log the commit information
					for _, commit := range commitHistory {
						logger.GlobalLogger.Verbosef("Failure %d - Related commit for %s: %s by %s at %s",
							index+1,
							failure.TestName,
							commit.Hash[:7],
							commit.Author,
							commit.Date.Format("2006-01-02"),
						)
						logger.GlobalLogger.Debugf("Failure %d - Commit message: %s", index+1, commit.Message)
						logger.GlobalLogger.Debugf("Failure %d - Files changed: %v", index+1, commit.Changes)
					}

					// Get line-specific changes if we have a line number
					if failure.LineNumber > 0 {
						// Get the exact line changes
						lineChanges, err := repo.GetLineChanges(relPath, failure.LineNumber)
						if err != nil {
							logger.GlobalLogger.Errorf("Failure %d - Failed to get line changes: %v", index+1, err)
						} else {
							logger.GlobalLogger.Debugf("Failure %d - Line changes:\n%s", index+1, lineChanges)
							failure.CodeChanges = lineChanges
						}

						// Get code context around the line-specific change
						absPath := filepath.Join(repo.Path(), relPath)
						context, err := repo.GetCodeContext(absPath, failure.LineNumber, contextLines)
						if err != nil {
							logger.GlobalLogger.Errorf("Failure %d - Failed to get code context: %v", index+1, err)
						} else {
							failure.Context.SurroundingCode = context
							logger.GlobalLogger.Debugf("Failure %d - Code context:\n%s", index+1, context)
						}

						// Get full file content
						// NOTE: Will be expensive if working with large files
						fullContent, err := repo.GetFullFileContent(absPath)
						if err != nil {
							logger.GlobalLogger.Errorf("Failure %d - Failed to get full file: %v", index+1, err)
						} else {
							failure.Context.FullFileContent = fullContent
						}

						// Get commits that modified this line
						lineCommits, err := repo.GetCommitsAffectingLines(relPath, []int{failure.LineNumber}, commitDepth)
						if err != nil {
							logger.GlobalLogger.Debugf("Failure %d - Failed to get line-specific commits: %v", index+1, err)
						} else {
							failure.RelatedCommits = lineCommits
							for _, commit := range lineCommits {
								logger.GlobalLogger.Verbosef("Failure %d - Line %d modified in commit %s: %s",
									index+1,
									failure.LineNumber,
									commit.Hash[:7],
									strings.Split(commit.Message, "\n")[0],
								)
							}
						}

					}
				}

				logger.GlobalLogger.Verbosef("Generating prompts for AI analysis")

				if batch {
					// Batch multiple test failures into one request
					prompt := ai.GenerateBatchPrompt(failures)
					logger.GlobalLogger.Debugf("Generated prompt for failure(s):\n%s", prompt)

					aiClient, err := ai.NewAIClient(aiOpts)
					if err != nil {
						logger.GlobalLogger.Errorf("Failed to create AI client: %v", err)
					}
					logger.GlobalLogger.Verbosef("Initialized AI Client for %s with %s", aiOpts.Provider, aiOpts.Model)

					aiResponse, err := aiClient.AnalyzeTestFailure(prompt)
					if err != nil {
						logger.GlobalLogger.Errorf("AI request failed: %v", err)
					} else {
						logger.GlobalLogger.Successf("AI response:\n%s", aiResponse)
					}
				} else {
					aiClient, err := ai.NewAIClient(aiOpts)
					if err != nil {
						logger.GlobalLogger.Errorf("Failed to create AI client: %v", err)
					}
					logger.GlobalLogger.Verbosef("Initialized AI Client for %s with %s", aiOpts.Provider, aiOpts.Model)

					// Generate prompt for each test failure
					for i, failure := range failures {
						prompt := ai.GeneratePrompt(failure)

						logger.GlobalLogger.Debugf("Generated prompt for failure %d:\n%s", i+1, prompt)

						aiResponse, err := aiClient.AnalyzeTestFailure(prompt)
						if err != nil {
							logger.GlobalLogger.Errorf("AI request failed: %v", err)
						} else {
							logger.GlobalLogger.Successf("AI response:\n%s", aiResponse)
						}
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
			Name: "Git options",
			Flags: []*pflag.Flag{
				cmd.Flags().Lookup("git-depth"),
				cmd.Flags().Lookup("context-lines"),
				cmd.Flags().Lookup("commit-depth"),
				cmd.Flags().Lookup("force"),
				cmd.Flags().Lookup("no-git"),
			},
		},
		{
			Name: "AI options",
			Flags: []*pflag.Flag{
				cmd.Flags().Lookup("api-key"),
				cmd.Flags().Lookup("model"),
				cmd.Flags().Lookup("ai-provider"),
				cmd.Flags().Lookup("batch"),
			},
		},
	}

	return groups
}

func getAIOptions(cmd *cobra.Command) (ai.AIOptions, error) {
	opts := ai.AIOptions{
		Provider: cmd.Flag("ai-provider").Value.String(),
		Model:    cmd.Flag("model").Value.String(),
	}

	// Get API key (flag takes precedence over env vars)
	apiKey, err := getAPIKey(cmd, opts.Provider)
	if err != nil {
		return ai.AIOptions{}, err
	}
	opts.APIKey = apiKey

	// Auto-detect provider if not specified
	if opts.Provider == "" {
		opts.Provider, err = detectProviderFromKey(opts.APIKey)
		if err != nil {
			return ai.AIOptions{}, err
		}
	} else {
		// Check for invalid provider provided
		if opts.Provider != "groq" && opts.Provider != "openai" {
			return ai.AIOptions{}, fmt.Errorf("Invalid ai provider: %s", opts.Provider)
		}
	}

	// Set default model if not specified
	if opts.Model == "" {
		opts.Model = getDefaultModel(opts.Provider)
	}

	return opts, nil
}

func getAPIKey(cmd *cobra.Command, provider string) (string, error) {
	if key := cmd.Flag("api-key").Value.String(); key != "" {
		return key, nil
	}

	switch provider {
	case "groq":
		return os.Getenv("GROQ_API_KEY"), nil
	case "openai":
		return os.Getenv("OPENAI_API_KEY"), nil
	default:
		if key := os.Getenv("GROQ_API_KEY"); key != "" {
			return key, nil
		}
		if key := os.Getenv("OPENAI_API_KEY"); key != "" {
			return key, nil
		}
	}

	return "", fmt.Errorf("No API key provided (use --api-key or set %s_API_KEY)", strings.ToUpper(provider))
}

func detectProviderFromKey(key string) (string, error) {
	switch {
	case strings.HasPrefix(key, "gsk_"):
		return "groq", nil
	case strings.HasPrefix(key, "sk-"):
		return "openai", nil
	default:
		return "", fmt.Errorf("Unable to detect provider from key format")
	}
}

func getDefaultModel(provider string) string {
	switch provider {
	case "groq":
		return "llama3-70b-8192"
	case "openai":
		return "gpt-3.5-turbo"
	default:
		return "llama3-70b-8192" // Fallback to Groq free model
	}
}
