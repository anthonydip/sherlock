package git

import (
	"os"
	"path/filepath"
	"strings"
)

func extractFilePath(location string) string {
	if len(location) >= 2 && location[1] == ':' {
		if lastColon := strings.LastIndex(location, ":"); lastColon > 1 {
			return location[:lastColon]
		}
		return location
	}
	return strings.Split(location, ":")[0]
}

func NormalizeTestPath(failureLocation string, repoPath string) (string, error) {
	path := extractFilePath(failureLocation)
	path = filepath.Clean(path)

	// Try to resolve path if already relative
	if !filepath.IsAbs(path) {
		// Try relative to test file directory first
		testDir := filepath.Dir(repoPath)
		possible := filepath.Join(testDir, path)
		if _, err := os.Stat(possible); err == nil {
			path = possible
		} else {
			// Try relative to repo root
			possible = filepath.Join(repoPath, path)
			if _, err := os.Stat(possible); err == nil {
				path = possible
			}
		}
	}

	// Ensure path is absolute
	if !filepath.IsAbs(path) {
		var err error
		path, err = filepath.Abs(filepath.Join(repoPath, path))
		if err != nil {
			return "", err
		}
	}

	// Check if the file exists
	if _, err := os.Stat(path); err != nil {
		// Attempt to locate filename within repo
		filename := filepath.Base(path)
		err := filepath.Walk(repoPath, func(walkPath string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && filepath.Base(walkPath) == filename {
				path = walkPath
				return filepath.SkipDir
			}
			return nil
		})
		if err != nil {
			return "", err
		}
	}

	relPath, err := filepath.Rel(repoPath, path)
	if err != nil {
		return "", err
	}

	return filepath.ToSlash(relPath), nil
}
