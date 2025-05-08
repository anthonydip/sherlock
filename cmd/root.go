package cmd

import (
	"fmt"

	"github.com/anthonydip/sherlock/cmd/analyze"
	"github.com/spf13/cobra"
)

type VersionInfo struct {
	Version   string
	BuildDate string
	GitCommit string
}

func NewRootCmd(versionInfo VersionInfo) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "sherlock",
		Short:   "AI-powered test failure analyzer",
		Long:    "Sherlock helps diagnose test failures by analyzing test outputs and git history",
		Version: formatVersion(versionInfo),
	}

	rootCmd.SetVersionTemplate("{{.Version}}")

	rootCmd.AddCommand(
		analyze.NewAnalyzeCmd(),
	)

	rootCmd.Flags().BoolP("version", "v", false, "Print version")

	return rootCmd
}

func formatVersion(versionInfo VersionInfo) string {
	return fmt.Sprintf("Sherlock v%s\nBuild Date: %s\nGit Commit: %s",
		versionInfo.Version,
		versionInfo.BuildDate,
		versionInfo.GitCommit)
}
