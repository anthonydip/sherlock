package cmd

import (
	"fmt"
	"os"

	"github.com/anthonydip/sherlock/cmd/analyze"
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
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "V", false, "Enable verbose output")

	rootCmd.Flags().BoolP("version", "v", false, "Print version")

	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		fmt.Fprintf(os.Stderr, "[ERROR] Invalid flag: %v\n\n", err)
		fmt.Fprintf(os.Stderr, "Run '%s --help' for usage\n", cmd.CommandPath())
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
