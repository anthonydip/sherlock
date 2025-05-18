package cmd

import (
	"fmt"
	"os"

	"github.com/anthonydip/sherlock/cmd/analyze"
	"github.com/anthonydip/sherlock/internal/cli"
	"github.com/anthonydip/sherlock/internal/logger"
	"github.com/spf13/cobra"
)

type VersionInfo struct {
	Version   string
	BuildDate string
	GitCommit string
}

func NewRootCmd(versionInfo VersionInfo) *cobra.Command {
	var (
		noColor bool
		verbose bool
		debug   bool
	)

	rootCmd := &cobra.Command{
		Use:     "sherlock",
		Short:   "AI-powered test failure analyzer",
		Long:    "Sherlock helps diagnose test failures by analyzing test outputs and git history",
		Version: formatVersion(versionInfo),
		Args: func(cmd *cobra.Command, args []string) error {
			// Reject any arguments passed to root command
			if len(args) > 0 {
				fmt.Fprintf(os.Stderr, "[ERROR] Unknown command %q\n\nRun 'sherlock --help' for available commands", args[0])
				return fmt.Errorf("Unknown command")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			// Show help when no subcommand is provided
			cmd.Help()
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			verbose, _ := cmd.Flags().GetBool("verbose")
			debug, _ := cmd.Flags().GetBool("debug")

			effectiveVerbose := verbose || debug

			logger.GlobalLogger = logger.New(effectiveVerbose, debug, !noColor)
			logger.GlobalLogger.Debugf("Initialized logger (verbose=%t, debug=%t, color=%t)", effectiveVerbose, debug, !noColor)
			logger.GlobalLogger.Debugf("Starting Sherlock %s", versionInfo.Version)
		},
	}

	rootCmd.SetVersionTemplate("{{.Version}}")

	// Root flags
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "V", false, "enable verbose output")

	rootCmd.Flags().BoolP("version", "v", false, "print version")

	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		option := cli.StripInvalidFlag(err)

		fmt.Fprintf(os.Stderr, "unknown option: %s\n", option)
		fmt.Fprintf(os.Stderr, "%s\n", cli.FormatRootUsage(cmd))
		return nil
	})

	rootCmd.AddCommand(
		analyze.NewAnalyzeCmd(),
	)

	return rootCmd
}

func formatVersion(versionInfo VersionInfo) string {
	return fmt.Sprintf("Sherlock %s\nBuild Date: %s\nGit Commit: %s",
		versionInfo.Version,
		versionInfo.BuildDate,
		versionInfo.GitCommit)
}
