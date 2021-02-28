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

var defaultExcludes []string

func glob(pattern string) string {
	return fmt.Sprintf("*%c%s", filepath.Separator, pattern)
}

func init() {
	defaultExcludes = []string{
		glob(".git"),
		glob(".github"),
		glob(".gitignore"),
		glob(".DS_Store"),
		glob(".vscode"),
	}
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

func Explore(sources, resources, excludes []string) (*Layout, error) {
	uniqueExcludes := make(map[string]struct{})
	for _, pattern := range excludes {
		uniqueExcludes[pattern] = struct{}{}
	}
	for _, pattern := range defaultExcludes {
		uniqueExcludes[pattern] = struct{}{}
	}
	excludeFilter := NewFilter(uniqueExcludes)

	uniqueResources := make(map[string]struct{})
	for _, path := range resources {
		uniqueResources[filepath.Clean(path)] = struct{}{}
	}
	resourcesFilter := NewFilter(uniqueResources)

	uniqueSources := make(map[string]struct{})
	for _, path := range sources {
		uniqueSources[filepath.Clean(path)] = struct{}{}
	}

	layout := &Layout{
		ExcludeFilter:  excludeFilter,
		ResourceFilter: resourcesFilter,
		Omitted:        make(map[string]error),
		ResourceDirs:   make(map[string]struct{}),
		ResourceFiles:  make(map[string]struct{}),
		SourceDirs:     make(map[string]struct{}),
		SourceFiles:    make(map[string]struct{}),
	}

	traverse(uniqueResources, func(path string, d fs.DirEntry, err error) {
		if err != nil {
			layout.Omitted[path] = fmt.Errorf("failed to traverse: %w", err)
			return
		}
		if excludeFilter.Match(path) {
			layout.Omitted[path] = nil
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
		if excludeFilter.Match(path) || resourcesFilter.Match(path) {
			layout.Omitted[path] = nil
			return
		}
		if d.IsDir() {
			layout.SourceDirs[path] = struct{}{}
		} else {
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
