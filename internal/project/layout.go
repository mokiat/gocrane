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

	"github.com/mokiat/gocrane/internal/filesystem"
	"github.com/mokiat/gocrane/internal/location"
)

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

type Summary struct {
	Errored map[string]error
	Omitted map[string]struct{}
	Visited map[string]struct{}

	WatchedDirs          map[filesystem.AbsolutePath]struct{}
	WatchedSourceFiles   map[filesystem.AbsolutePath]struct{}
	WatchedResourceFiles map[filesystem.AbsolutePath]struct{}
}

type Layout struct {
	Errored        map[string]error
	Ignored        map[string]struct{}
	WatchDirs      []string
	WatchFilter    *filesystem.FilterTree
	SourceFiles    []string
	SourceFilter   location.Filter
	ResourceFilter location.Filter
}

func (l *Layout) PrintToLog() {
	log.Printf("encountered an error with %d files or folders", len(l.Errored))
	for file, err := range l.Errored {
		log.Printf("errored: %s (%s)", file, err)
	}

	log.Printf("omitted %d files or folders", len(l.Ignored))
	for file := range l.Ignored {
		log.Printf("omitted: %s", file)
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

func Explore(dirs, dirExcludes, sources, sourceExludes, resources, resourceExcludes []string) *Layout {
	errored := make(map[string]error)
	omitted := make(map[string]struct{})

	watchFilter := filesystem.NewFilterTree()
	for _, dir := range dirs {
		if filesystem.IsGlob(dir) {
			watchFilter.AcceptGlob(dir)
		} else {
			path, err := filesystem.ToAbsolutePath(dir)
			if err != nil {
				errored[path] = fmt.Errorf("failed to convert path to absolute: %w", err)
				continue
			}
			watchFilter.AcceptPath(path)
		}
	}
	for _, dir := range dirExcludes {
		if filesystem.IsGlob(dir) {
			watchFilter.RejectGlob(dir)
		} else {
			path, err := filesystem.ToAbsolutePath(dir)
			if err != nil {
				errored[path] = fmt.Errorf("failed to convert path to absolute: %w", err)
				continue
			}
			watchFilter.RejectPath(path)
		}
	}

	// watchFilter := location.NotFilter(
	// 	buildFilter(dirExcludes, errored),
	// )
	sourcesFilter := location.AndFilter(
		buildFilter(sources, errored),
		location.NotFilter(
			buildFilter(sourceExludes, errored),
		),
	)
	resourcesFilter := location.AndFilter(
		buildFilter(resources, errored),
		location.NotFilter(
			buildFilter(resourceExcludes, errored),
		),
	)

	uniqueDirs := make(map[string]struct{})
	uniqueFiles := make(map[string]struct{})
	for _, dir := range dirs {
		result := location.Traverse(dir, watchFilter, func(path filesystem.AbsolutePath, isDir bool) error {
			if isDir {
				uniqueDirs[path] = struct{}{}
			} else {
				uniqueFiles[path] = struct{}{}
			}
			return nil
		})
		for path := range result.IgnoredPaths {
			omitted[path] = struct{}{}
		}
		for path, err := range result.ErroredPaths {
			errored[path] = err
		}
	}

	watchDirs := make([]string, 0, len(uniqueDirs))
	for path := range uniqueDirs {
		watchDirs = append(watchDirs, path)
	}
	sort.Strings(watchDirs)

	sourceFiles := make([]string, 0, len(uniqueFiles))
	for path := range uniqueFiles {
		if sourcesFilter.Match(path) {
			sourceFiles = append(sourceFiles, path)
		}
	}
	sort.Strings(sourceFiles)

	return &Layout{
		Errored:        errored,
		Ignored:        omitted,
		WatchDirs:      watchDirs,
		WatchFilter:    watchFilter,
		SourceFiles:    sourceFiles,
		SourceFilter:   sourcesFilter,
		ResourceFilter: resourcesFilter,
	}
}

func buildFilter(targets []string, errored map[string]error) location.Filter {
	var filters []location.Filter
	for _, target := range targets {
		if filesystem.IsGlob(target) {
			filters = append(filters, location.GlobFilter(target))
		} else {
			path, err := filepath.Abs(target)
			if err != nil {
				errored[target] = fmt.Errorf("failed to convert path to absolute: %w", err)
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
