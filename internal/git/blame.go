package git

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func (r *Repository) GetBlame(path string) (*git.BlameResult, error) {
	head, err := r.repo.Head()
	if err != nil {
		return nil, err
	}

	commit, err := r.repo.CommitObject(head.Hash())
	if err != nil {
		return nil, err
	}

	return git.Blame(commit, path)
}

func (r *Repository) GetCommitsAffectingLines(path string, lines []int, limit int) ([]CommitInfo, error) {
	blame, err := r.GetBlame(path)
	if err != nil {
		return nil, err
	}

	var commits []CommitInfo
	seen := make(map[string]bool)

	for _, line := range lines {
		if line <= 0 || line > len(blame.Lines) {
			continue
		}

		commitHash := blame.Lines[line-1].Hash.String()
		if seen[commitHash] {
			continue
		}
		seen[commitHash] = true

		commit, err := r.repo.CommitObject(blame.Lines[line-1].Hash)
		if err != nil {
			continue
		}

		fileIter, err := commit.Files()
		if err != nil {
			continue
		}

		var changes []string
		fileIter.ForEach(func(f *object.File) error {
			changes = append(changes, f.Name)
			return nil
		})

		commits = append(commits, CommitInfo{
			Hash:    commitHash,
			Author:  commit.Author.String(),
			Date:    commit.Author.When,
			Message: strings.TrimSpace(commit.Message),
			Changes: changes,
		})

		if len(commits) >= limit {
			break
		}
	}

	return commits, nil
}

func (r *Repository) GetLineChanges(path string, line int) (string, error) {
	blame, err := r.GetBlame(path)
	if err != nil {
		return "", err
	}

	if line <= 0 || line > len(blame.Lines) {
		return "", fmt.Errorf("line out of range")
	}

	commit, err := r.repo.CommitObject(blame.Lines[line-1].Hash)
	if err != nil {
		return "", err
	}

	// Handle initial commit or file creation case
	if commit.NumParents() == 0 {
		return fmt.Sprintf("+ %s (file created in this commit)", blame.Lines[line-1].Text), nil
	}

	parent, err := commit.Parent(0)
	if err != nil {
		return "", err
	}

	// Check if file was added in this commit
	_, err = parent.File(path)
	if err == object.ErrFileNotFound {
		return fmt.Sprintf("+ %s (file added in this commit)", blame.Lines[line-1].Text), nil
	} else if err != nil {
		return "", err
	}

	// Get diff for modified files
	patch, err := parent.Patch(commit)
	if err != nil {
		return "", err
	}

	var changes strings.Builder
	targetLine := blame.Lines[line-1]

	for _, filePatch := range patch.FilePatches() {
		_, to := filePatch.Files()
		if to == nil || to.Path() != path {
			continue
		}

		// Get line numbers from chunk headers
		for _, chunk := range filePatch.Chunks() {
			if strings.Contains(chunk.Content(), targetLine.Text) {
				switch chunk.Type() {
				case diff.Delete:
					changes.WriteString(fmt.Sprintf("- %s\n", strings.TrimSpace(chunk.Content())))
				case diff.Add:
					changes.WriteString(fmt.Sprintf("+ %s\n", strings.TrimSpace(chunk.Content())))
				case diff.Equal:
					changes.WriteString(fmt.Sprintf("  %s\n", strings.TrimSpace(chunk.Content())))
				}
			}
		}
	}

	if changes.Len() == 0 {
		return fmt.Sprintf("  %s (no changes in this commit)", targetLine.Text), nil
	}

	return changes.String(), nil
}
