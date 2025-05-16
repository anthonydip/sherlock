package main

import (
	"os"

	"github.com/anthonydip/sherlock/cmd"
)

var (
	version   = "dev"
	buildDate = "unset"
	gitCommit = "uncommitted"
)

func main() {
	versionInfo := cmd.VersionInfo{
		Version:   version,
		BuildDate: buildDate,
		GitCommit: gitCommit,
	}

	rootCmd := cmd.NewRootCmd(versionInfo)
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
