package git

import (
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
