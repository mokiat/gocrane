package project

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/mokiat/gocrane/internal/filesystem"
)

// Analyze traverses the folders specified by rootDirs and evaluates which
// files and folders would be watched based on the specified filters.
//
// The outcome of the analysis is returned as a Summary.
func Analyze(rootDirs []filesystem.AbsolutePath, watchFilter, sourceFilter, resourceFilter *filesystem.FilterTree) *Summary {
	var (
		errored = make(map[string]error)
		omitted = make(map[string]struct{})
		visited = make(map[string]struct{})

		watchedDirs          = make(map[filesystem.AbsolutePath]struct{})
		watchedFiles         = make(map[filesystem.AbsolutePath]struct{})
		watchedSourceFiles   = make(map[filesystem.AbsolutePath]struct{})
		watchedResourceFiles = make(map[filesystem.AbsolutePath]struct{})
	)

	for _, root := range rootDirs {
		filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				errored[p] = fmt.Errorf("error traversing path: %w", err)
				return filepath.SkipDir
			}
			absPath, err := filesystem.ToAbsolutePath(p)
			if err != nil {
				errored[p] = fmt.Errorf("error converting path to absolute: %w", err)
				return filepath.SkipDir
			}
			if !watchFilter.IsAccepted(absPath) {
				omitted[p] = struct{}{}
				return filepath.SkipDir
			}
			if d.IsDir() {
				watchedDirs[absPath] = struct{}{}
			} else {
				watchedFiles[absPath] = struct{}{}
			}
			visited[p] = struct{}{}
			return nil
		})
	}

	for absPath := range watchedFiles {
		if sourceFilter.IsAccepted(absPath) {
			watchedSourceFiles[absPath] = struct{}{}
		}
		if resourceFilter.IsAccepted(absPath) {
			watchedResourceFiles[absPath] = struct{}{}
		}
	}

	return &Summary{
		Errored: errored,
		Omitted: omitted,
		Visited: visited,

		WatchedDirs:          watchedDirs,
		WatchedSourceFiles:   watchedSourceFiles,
		WatchedResourceFiles: watchedResourceFiles,
	}
}

// Summary is the outcome of a project analysis.
type Summary struct {
	Errored map[string]error
	Omitted map[string]struct{}
	Visited map[string]struct{}

	WatchedDirs          map[filesystem.AbsolutePath]struct{}
	WatchedSourceFiles   map[filesystem.AbsolutePath]struct{}
	WatchedResourceFiles map[filesystem.AbsolutePath]struct{}
}
