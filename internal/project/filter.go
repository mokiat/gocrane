package project

import (
	"fmt"
	"path/filepath"
	"strings"
)

func Glob(pattern string) string {
	return fmt.Sprintf("*%c%s", filepath.Separator, pattern)
}

func NewFilter(patterns []string) *Filter {
	globPrefix := fmt.Sprintf("*%c", filepath.Separator)
	paths := make(map[string]struct{})
	globs := make(map[string]struct{})
	for _, pattern := range patterns {
		if strings.HasPrefix(pattern, globPrefix) {
			globs[strings.TrimPrefix(pattern, globPrefix)] = struct{}{}
		} else {
			paths[filepath.Clean(pattern)] = struct{}{}
		}
	}
	return &Filter{
		paths: paths,
		globs: globs,
	}
}

type Filter struct {
	paths map[string]struct{}
	globs map[string]struct{}
}

func (f *Filter) Empty() bool {
	return len(f.paths) == 0 && len(f.globs) == 0
}

func (f *Filter) Paths() map[string]struct{} {
	return f.paths
}

func (f *Filter) Globs() map[string]struct{} {
	return f.globs
}

func (f *Filter) Match(file string) bool {
	for path := range f.paths {
		if strings.HasPrefix(file, path) {
			return true
		}
	}
	segments := strings.Split(file, string(filepath.Separator))
	for glob := range f.globs {
		for _, segment := range segments {
			match, err := filepath.Match(glob, segment)
			if err == nil && match {
				return true
			}
		}
	}
	return false
}

func (f *Filter) MatchAll(files ...string) bool {
	for _, file := range files {
		if !f.Match(file) {
			return false
		}
	}
	return true
}
