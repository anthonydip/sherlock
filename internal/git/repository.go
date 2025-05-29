package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthonydip/sherlock/internal/logger"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Repository struct {
	path string
	repo *git.Repository
}

// Open Git repository at the given path
func OpenRepository(path string, depth int) (*Repository, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	repoPath, err := findRepositoryRoot(absPath, depth)
	if err != nil {
		return nil, err
	}

	logger.GlobalLogger.Verbosef("Opening Git repository at: %s", repoPath)

	// Open with go-git
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		if errors.Is(err, git.ErrRepositoryNotExists) {
			return nil, ErrNotAGitRepository
		}
		return nil, err
	}

	return &Repository{
		path: repoPath,
		repo: repo,
	}, nil
}

func findRepositoryRoot(startPath string, depth int) (string, error) {
	current := startPath

	logger.GlobalLogger.Debugf("Searching for Git repository root with maximum depth of %d", depth)

	for i := 0; i < depth; i++ {
		gitPath := filepath.Join(current, ".git")

		logger.GlobalLogger.Debugf("Checking for Git repository at: %s", gitPath)

		// Check if .git exists
		if fi, err := os.Stat(gitPath); err == nil {
			if fi.IsDir() {
				return current, nil
			}

			// Handle git submodules
			if content, err := os.ReadFile(gitPath); err == nil {
				if strings.HasPrefix(string(content), "gitdir: ") {
					return current, nil
				}
			}
		}

		// Move up one directory
		parent := filepath.Dir(current)
		if parent == current {
			break // Reached filesystem root
		}
		current = parent
	}

	return "", ErrNotAGitRepository
}

func (r *Repository) Path() string {
	return r.path
}

// Check for uncomitted changes in working tree
func (r *Repository) IsDirty() (bool, error) {
	logger.GlobalLogger.Verbosef("Checking repository status at: %s", r.path)

	w, err := r.repo.Worktree()
	if err != nil {
		return false, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := w.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get git status: %w", err)
	}

	logger.GlobalLogger.Debugf("Git status entries: %d", len(status))

	hasChanges := false
	for path, fileStatus := range status {
		wtCode := string(fileStatus.Worktree)
		stCode := string(fileStatus.Staging)

		logger.GlobalLogger.Debugf("File status: %s (Worktree: %s, Staging: %s)",
			path, wtCode, stCode)

		// Skip files that are only untracked
		if wtCode == "?" && stCode == " " {
			logger.GlobalLogger.Debugf("Ignoring untracked file: %s", path)
			continue
		}

		// Skip files that only have mode changes
		if wtCode == " " && stCode == "M" {
			logger.GlobalLogger.Debugf("Ignoring mode-only change: %s", path)
			continue
		}

		// Handle potential line ending changes
		if wtCode == "M" && stCode == " " {
			contentChanged, err := r.hasActualContentChanges(path)
			if err != nil {
				logger.GlobalLogger.Debugf("Failed to verify content changes for %s: %v", path, err)
				hasChanges = true
				continue
			}
			if contentChanged {
				logger.GlobalLogger.Debugf("Verified content changes: %s", path)
				hasChanges = true
			} else {
				logger.GlobalLogger.Debugf("No actual content changes: %s", path)
			}
			continue
		}

		// All other modification states
		if fileStatus.Worktree != git.Unmodified || fileStatus.Staging != git.Unmodified {
			logger.GlobalLogger.Debugf("Significant modification detected: %s", path)
			hasChanges = true
		}
	}

	if hasChanges {
		logger.GlobalLogger.Verbosef("Repository is dirty, uncommitted changes detected")
	} else {
		logger.GlobalLogger.Verbosef("Repository is clean, no uncommitted changes")
	}

	return hasChanges, nil
}

func (r *Repository) hasActualContentChanges(path string) (bool, error) {
	// Get HEAD commit
	head, err := r.repo.Head()
	if err != nil {
		return false, err
	}

	commit, err := r.repo.CommitObject(head.Hash())
	if err != nil {
		return false, err
	}

	// Get file from HEAD
	headFile, err := commit.File(path)
	if err == object.ErrFileNotFound {
		// File is newly created
		return true, nil
	} else if err != nil {
		return false, err
	}

	// Get HEAD file content
	headContent, err := headFile.Contents()
	if err != nil {
		return false, err
	}

	// Get current file content
	currentContent, err := os.ReadFile(filepath.Join(r.path, path))
	if err != nil {
		return false, err
	}

	// Normalize line endings for comparison
	normalizedHead := strings.ReplaceAll(headContent, "\r\n", "\n")
	normalizedCurrent := strings.ReplaceAll(string(currentContent), "\r\n", "\n")

	// Compare normalized content
	if normalizedHead == normalizedCurrent {
		logger.GlobalLogger.Debugf("No actual content difference after normalization: %s", path)
		return false, nil
	}

	logger.GlobalLogger.Debugf("Content differences found in: %s", path)
	return true, nil
}
