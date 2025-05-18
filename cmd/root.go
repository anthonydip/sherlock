package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/anthonydip/sherlock/cmd/analyze"
	"github.com/anthonydip/sherlock/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
		var option string

		switch {
		case strings.Contains(err.Error(), "unknown shorthand flag"):
			parts := strings.Split(err.Error(), "'")
			if len(parts) > 1 {
				option = "-" + parts[1]
			}
		case strings.Contains(err.Error(), "unknown flag"):
			parts := strings.Split(err.Error(), " ")
			if len(parts) > 2 {
				option = parts[2]
			}
		default:
			option = err.Error()
		}

		fmt.Fprintf(os.Stderr, "unknown option: %s\n", option)
		fmt.Fprintf(os.Stderr, "%s\n", formatStyleUsage(cmd))
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

func formatStyleUsage(cmd *cobra.Command) string {
	usage := fmt.Sprintf("usage: %s", cmd.CommandPath())
	padding := strings.Repeat(" ", len(usage)+1)

	var flagGroups []string
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		if flag.Shorthand != "" {
			flagGroups = append(flagGroups, fmt.Sprintf("[-%s | --%s]", flag.Shorthand, flag.Name))
		} else {
			flagGroups = append(flagGroups, fmt.Sprintf("[--%s]", flag.Name))
		}
	})

	// Build lines ensuring no flag group is split
	maxWidth := 80
	currentLine := usage + " "
	var lines []string

	for _, group := range flagGroups {
		if len(currentLine)+len(group)+1 > maxWidth {
			// If we're not on the first line, add padding
			if len(lines) > 0 {
				currentLine = padding + currentLine[len(usage)+1:]
			}
			lines = append(lines, currentLine)
			currentLine = padding + group
		} else {
			if currentLine != usage+" " {
				currentLine += " "
			}
			currentLine += group
		}
	}

	commandPart := "<command> [<args>]"
	if len(currentLine)+1+len(commandPart) <= maxWidth {
		currentLine += " " + commandPart
	} else {
		if currentLine != "" {
			if len(lines) > 0 {
				currentLine = padding + currentLine[len(usage)+1:]
			}
			lines = append(lines, currentLine)
		}
		currentLine = padding + commandPart
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return strings.Join(lines, "\n")
}
