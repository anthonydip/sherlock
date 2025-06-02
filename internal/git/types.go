package git

import (
	"time"

	"github.com/go-git/go-git/v5"
)

type Repository struct {
	path string
	repo *git.Repository
}

type CommitInfo struct {
	Hash    string
	Author  string
	Date    time.Time
	Message string
	Changes []string
}
