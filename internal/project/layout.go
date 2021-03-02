package project

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

func CombineFileSets(a, b FileSet) FileSet {
	result := make(FileSet, len(a)+len(b))
	for k, v := range a {
		result[k] = v
	}
	for k, v := range b {
		result[k] = v
	}
	return result
}

type FileSet map[string]struct{}

func (s FileSet) SortedList() []string {
	result := make([]string, 0, len(s))
	for file := range s {
		result = append(result, file)
	}
	sort.Strings(result)
	return result
}

type Layout struct {
	Omitted        map[string]error
	ExcludeFilter  *Filter
	ResourceFiles  FileSet
	ResourceDirs   FileSet
	ResourceFilter *Filter
	SourceFiles    FileSet
	SourceDirs     FileSet
}

func (l *Layout) Digest() (string, error) {
	dig := sha256.New()
	for _, file := range l.SourceFiles.SortedList() {
		if err := writeFileDigest(file, dig); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%x", dig.Sum(nil)), nil
}

func Explore(sources []string, resourcesFilter, includesFilter, excludesFilter *Filter) (*Layout, error) {
	uniqueSources := make(map[string]struct{})
	for _, path := range sources {
		uniqueSources[filepath.Clean(path)] = struct{}{}
	}

	layout := &Layout{
		ExcludeFilter:  excludesFilter,
		ResourceFilter: resourcesFilter,
		Omitted:        make(map[string]error),
		ResourceDirs:   make(map[string]struct{}),
		ResourceFiles:  make(map[string]struct{}),
		SourceDirs:     make(map[string]struct{}),
		SourceFiles:    make(map[string]struct{}),
	}

	traverse(resourcesFilter.Paths(), func(path string, d fs.DirEntry, err error) {
		if err != nil {
			layout.Omitted[path] = fmt.Errorf("failed to traverse: %w", err)
			return
		}
		if excludesFilter.Match(path) {
			layout.Omitted[path] = fmt.Errorf("is excluded")
			return
		}
		if d.IsDir() {
			layout.ResourceDirs[path] = struct{}{}
		} else {
			layout.ResourceFiles[path] = struct{}{}
		}
	})

	traverse(uniqueSources, func(path string, d fs.DirEntry, err error) {
		if err != nil {
			layout.Omitted[path] = fmt.Errorf("failed to traverse: %w", err)
			return
		}
		if resourcesFilter.Match(path) {
			return
		}
		if excludesFilter.Match(path) {
			layout.Omitted[path] = fmt.Errorf("is excluded")
			return
		}
		if d.IsDir() {
			layout.SourceDirs[path] = struct{}{}
		} else {
			if !includesFilter.Empty() && !includesFilter.Match(path) {
				layout.Omitted[path] = fmt.Errorf("not included")
				return
			}
			layout.SourceFiles[path] = struct{}{}
		}
	})

	return layout, nil
}

type traverseFunc func(path string, d fs.DirEntry, err error)

func traverse(roots map[string]struct{}, fn traverseFunc) {
	for root := range roots {
		filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			path = filepath.Clean(path)
			fn(path, d, err)
			if err != nil {
				return filepath.SkipDir
			}
			return nil
		})
	}
}

func writeFileDigest(file string, h hash.Hash) error {
	stat, err := os.Stat(file)
	if err != nil {
		return fmt.Errorf("failed to state file %q: %w", file, err)
	}
	// Note: Don't include millisecond precision, as that seems to differ between
	// host and client machine (in some cases it is not included).
	const timeFormat = "2006/01/02 15:04:05"
	fmt.Fprint(h, len(file), file, stat.ModTime().UTC().Format(timeFormat), stat.Size())
	return nil
}
