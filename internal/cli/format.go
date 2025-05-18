package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type FlagGroup struct {
	Name  string
	Flags []*pflag.Flag
}

func StripInvalidFlag(err error) string {
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

	return option
}

func FormatRootUsage(cmd *cobra.Command) string {
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

func FormatSubcommandUsage(cmd *cobra.Command, groups []FlagGroup) string {
	var builder strings.Builder

	// Usage line
	builder.WriteString(fmt.Sprintf("usage: %s [<options>] <test-output>\n\n", cmd.CommandPath()))

	// Flag groups
	for _, group := range groups {
		builder.WriteString(group.Name + ":\n")
		for _, flag := range group.Flags {
			if flag.Hidden {
				continue
			}

			line := "    "
			if flag.Shorthand != "" {
				line += fmt.Sprintf("-%s, ", flag.Shorthand)
			}
			line += fmt.Sprintf("--%s", flag.Name)

			// Add value placeholder if not boolean
			if flag.Value.Type() != "bool" {
				line += fmt.Sprintf(" <%s>", flag.Name)
			}

			// Align descriptions at 40 characters
			if len(line) < 40 {
				line += strings.Repeat(" ", 40-len(line))
			} else {
				line += "\n    " + strings.Repeat(" ", 40)
			}

			line += flag.Usage + "\n"
			builder.WriteString(line)
		}
		builder.WriteString("\n")
	}

	return builder.String()
}
