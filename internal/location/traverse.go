package location

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/mokiat/gocrane/internal/filesystem"
)

var ErrSkip = fmt.Errorf("skipping path")

type TraversalFunc func(path filesystem.AbsolutePath, isDir bool) error

type TraversalResult struct {
	VisitedPaths map[string]struct{}
	ErroredPaths map[string]error
	IgnoredPaths map[string]struct{}
}

func Traverse(root string, watchTree *filesystem.FilterTree, fn TraversalFunc) TraversalResult {
	result := TraversalResult{
		VisitedPaths: make(map[string]struct{}),
		ErroredPaths: make(map[string]error),
		IgnoredPaths: make(map[string]struct{}),
	}
	filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			result.ErroredPaths[p] = fmt.Errorf("failed to traverse path: %w", err)
			return filepath.SkipDir
		}
		path, err := filesystem.ToAbsolutePath(p)
		if err != nil {
			result.ErroredPaths[p] = fmt.Errorf("failed to process path: %w", err)
			return filepath.SkipDir
		}
		if !watchTree.IsAccepted(path) {
			result.IgnoredPaths[p] = struct{}{}
			return filepath.SkipDir
		}
		if err := fn(path, d.IsDir()); err != nil {
			if err == ErrSkip {
				result.IgnoredPaths[path] = struct{}{}
			} else {
				result.ErroredPaths[path] = err
			}
			return filepath.SkipDir
		}
		result.VisitedPaths[path] = struct{}{}
		return nil
	})
	return result
}
