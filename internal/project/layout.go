package project

import (
	"fmt"
	"io/fs"
	"path/filepath"
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

type Layout struct {
	Omitted        map[string]error
	ExcludeFilter  *Filter
	ResourceFiles  map[string]struct{}
	ResourceDirs   map[string]struct{}
	ResourceFilter *Filter
	SourceFiles    map[string]struct{}
	SourceDirs     map[string]struct{}
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
