package git

import (
	"fmt"
	"os"
	"strings"
)

func (r *Repository) GetCodeContext(path string, lineNum int, contextLines int) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	start := max(0, lineNum-1-contextLines) // lineNum is 1-based
	end := min(len(lines), lineNum-1+contextLines)

	var builder strings.Builder
	for i := start; i <= end; i++ {
		prefix := "  "
		if i == lineNum-1 {
			prefix = ">>"
		}
		builder.WriteString(fmt.Sprintf("%s L%d: %s\n", prefix, i+1, lines[i]))
	}
	return builder.String(), nil
}

func (r *Repository) GetFullFileContent(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
