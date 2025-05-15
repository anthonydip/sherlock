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

	if err := rootCmd.Execute(); err != nil {
		// fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
