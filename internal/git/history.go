package git

import (
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func (r *Repository) GetFileHistory(path string, limit int) ([]*object.Commit, error) {
	head, err := r.repo.Head()
	if err != nil {
		return nil, err
	}

	commitIter, err := r.repo.Log(&git.LogOptions{
		From:  head.Hash(),
		Order: git.LogOrderCommitterTime,
		PathFilter: func(p string) bool {
			return p == path
		},
	})
	if err != nil {
		return nil, err
	}

	var commits []*object.Commit
	err = commitIter.ForEach(func(c *object.Commit) error {
		if len(commits) >= limit {
			return nil
		}
		commits = append(commits, c)
		return nil
	})

	return commits, err
}

func (r *Repository) GetEnhancedFileHistory(path string, limit int) ([]CommitInfo, error) {
	commits, err := r.GetFileHistory(path, limit)
	if err != nil {
		return nil, err
	}

	var commitInfos []CommitInfo
	for _, commit := range commits {
		fileIter, err := commit.Files()
		if err != nil {
			continue
		}

		var changes []string
		fileIter.ForEach(func(f *object.File) error {
			changes = append(changes, f.Name)
			return nil
		})

		commitInfos = append(commitInfos, CommitInfo{
			Hash:    commit.Hash.String(),
			Author:  commit.Author.String(),
			Date:    commit.Author.When,
			Message: strings.TrimSpace(commit.Message),
			Changes: changes,
		})
	}

	return commitInfos, nil
}
