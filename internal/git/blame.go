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

	// Get the diff for this commit
	var parent *object.Commit
	if commit.NumParents() > 0 {
		parent, err = commit.Parent(0)
		if err != nil {
			return "", err
		}
	} else {
		// Handle initial commit case
		return blame.Lines[line-1].Text, nil
	}

	patch, err := parent.Patch(commit)
	if err != nil {
		return "", err
	}

	// Find changes affecting our line
	var changes strings.Builder
	targetLine := strings.TrimSpace(blame.Lines[line-1].Text)

	for _, filePatch := range patch.FilePatches() {
		if filePatch.IsBinary() {
			continue
		}

		_, to := filePatch.Files()
		if to == nil || to.Path() != path {
			continue
		}

		for _, chunk := range filePatch.Chunks() {
			chunkContent := strings.TrimSpace(chunk.Content())
			if chunkContent == targetLine {
				// Determine operation type
				switch chunk.Type() {
				case diff.Delete:
					changes.WriteString(fmt.Sprintf("- %s\n", chunkContent))
				case diff.Add:
					changes.WriteString(fmt.Sprintf("+ %s\n", chunkContent))
				case diff.Equal:
					changes.WriteString(fmt.Sprintf("  %s\n", chunkContent))
				}
			}
		}
	}

	if changes.Len() == 0 {
		return blame.Lines[line-1].Text, nil
	}

	return changes.String(), nil
}
