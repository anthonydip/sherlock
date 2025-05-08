package git

import (
	"github.com/go-git/go-git/v5"
)

func GetDiff(repoPath string, sinceRef string) (string, error) {
	repo, err := git.PlainOpen(repoPath)

	return "", nil
}
