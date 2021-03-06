package project

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/mokiat/gocrane/internal/location"
)

type Layout struct {
	Omitted         map[string]error
	WatchDirs       []string
	SourceFiles     []string
	IncludeFilter   location.Filter
	ExcludeFilter   location.Filter
	SourcesFilter   location.Filter
	ResourcesFilter location.Filter
}

func (l *Layout) PrintToLog() {
	log.Printf("omitted %d files or folders", len(l.Omitted))
	for file, err := range l.Omitted {
		log.Printf("omitted: %s (%s)", file, err)
	}

	log.Printf("found %d directories to watch", len(l.WatchDirs))
	for _, dir := range l.WatchDirs {
		log.Printf("watch dir: %s", dir)
	}

	log.Printf("found %d files to use for digest", len(l.SourceFiles))
	for _, file := range l.SourceFiles {
		log.Printf("source file: %s", file)
	}
}

func (l *Layout) Digest() (string, error) {
	dig := sha256.New()
	for _, file := range l.SourceFiles {
		if err := writeFileDigest(string(file), dig); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%x", dig.Sum(nil)), nil
}

func Explore(includes, excludes, sources, resources []string) *Layout {
	var omitted map[string]error

	includeFilter := buildFilter(includes, omitted)
	excludeFilter := buildFilter(excludes, omitted)
	sourcesFilter := buildFilter(sources, omitted)
	resourcesFilter := buildFilter(resources, omitted)

	uniqueDirs := make(map[string]struct{})
	uniqueFiles := make(map[string]struct{})
	consider := func(p string, isDir bool) bool {
		path, err := filepath.Abs(p)
		if err != nil {
			omitted[p] = fmt.Errorf("failed to convert path to absolute %q: %w", p, err)
			return false // don't traverse children
		}
		if excludeFilter.Match(path) {
			omitted[p] = fmt.Errorf("path is excluded")
			return false // don't traverse children
		}
		if isDir {
			uniqueDirs[path] = struct{}{}
		} else {
			uniqueFiles[path] = struct{}{}
		}
		return true
	}
	traverse(includes, omitted, func(p string, d fs.DirEntry) bool {
		if d.IsDir() {
			return consider(p, true)
		} else {
			consider(filepath.Dir(p), true)
			return consider(p, false)
		}
	})

	watchDirs := make([]string, 0, len(uniqueDirs))
	for path := range uniqueDirs {
		watchDirs = append(watchDirs, path)
	}
	sort.Strings(watchDirs)

	sourceFiles := make([]string, 0, len(uniqueFiles))
	for path := range uniqueFiles {
		if includeFilter.Match(path) && sourcesFilter.Match(path) {
			sourceFiles = append(sourceFiles, path)
		}
	}
	sort.Strings(sourceFiles)

	return &Layout{
		WatchDirs:       watchDirs,
		SourceFiles:     sourceFiles,
		IncludeFilter:   includeFilter,
		ExcludeFilter:   excludeFilter,
		SourcesFilter:   sourcesFilter,
		ResourcesFilter: resourcesFilter,
	}
}

type traverseFunc func(p string, d fs.DirEntry) bool

func traverse(roots []string, omitted map[string]error, fn traverseFunc) {
	for _, root := range roots {
		filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				omitted[p] = fmt.Errorf("failed to traverse: %w", err)
				return filepath.SkipDir
			}
			if !fn(p, d) {
				return filepath.SkipDir
			}
			return nil
		})
	}
}

func buildFilter(targets []string, omitted map[string]error) location.Filter {
	var filters []location.Filter
	for _, target := range targets {
		if location.AppearsGlob(target) {
			filters = append(filters, location.GlobFilter(target))
		} else {
			path, err := filepath.Abs(target)
			if err != nil {
				omitted[target] = fmt.Errorf("failed to convert path to absolute: %w", err)
			} else {
				filters = append(filters, location.PathFilter(path))
			}
		}
	}
	return location.OrFilter(filters...)
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
