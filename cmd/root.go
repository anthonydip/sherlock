package cmd

import (
	"fmt"

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
		debug   bool
	)

	rootCmd := &cobra.Command{
		Use:     "sherlock",
		Short:   "AI-powered test failure analyzer",
		Long:    "Sherlock helps diagnose test failures by analyzing test outputs and git history",
		Version: formatVersion(versionInfo),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logger.GlobalLogger = logger.New(false, debug, !noColor)
		},
	}

	rootCmd.SetVersionTemplate("{{.Version}}")

	// Root flags
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug output")

	rootCmd.Flags().BoolP("version", "v", false, "Print version")

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
