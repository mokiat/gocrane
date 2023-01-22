package filesystem

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// ErrSkip can be returned to indicate that the current file or folder
// should be skipped from any further traversal.
var ErrSkip = fmt.Errorf("skip file or folder")

// TraverseFunc is called for each visited folder.
type TraverseFunc func(path string, isDir bool, err error) error

// Traverse attempts to simplify file traversal.
//
// The underlying filepath.WalkDir only works well for directories as
// root elements. Passing a file as root changes the contract substantially
// making it hard on the client.
func Traverse(root string, callback TraverseFunc) {
	// Handle the case where the root path is a file.
	info, err := os.Lstat(root)
	if err != nil {
		callback(root, false, fmt.Errorf("error getting info on root path %q: %w", root, err))
	}
	if !info.IsDir() {
		callback(root, info.IsDir(), nil)
		return
	}

	// The root is a dir so we can use filepath.WalkDir now.
	filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		// The only case where `d` is nil is when the root is a file. Since we
		// have covered this case already, we can use it at will.
		if cbErr := callback(p, d.IsDir(), err); cbErr != nil || err != nil {
			if d.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}
		return nil
	})
}
