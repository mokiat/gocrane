package command

import (
	"fmt"

	"github.com/mokiat/gocrane/internal/filesystem"
)

func buildFilterTree(accepted, rejected []string) (*filesystem.FilterTree, error) {
	result := filesystem.NewFilterTree()
	for _, entry := range accepted {
		if filesystem.IsGlob(entry) {
			result.AcceptGlob(entry)
		} else {
			path, err := filesystem.ToAbsolutePath(entry)
			if err != nil {
				return nil, fmt.Errorf("error processing accept rule: %w", err)
			}
			result.AcceptPath(path)
		}
	}
	for _, entry := range rejected {
		if filesystem.IsGlob(entry) {
			result.RejectGlob(entry)
		} else {
			path, err := filesystem.ToAbsolutePath(entry)
			if err != nil {
				return nil, fmt.Errorf("error processing reject rule: %w", err)
			}
			result.RejectPath(path)
		}
	}
	return result, nil
}
