package command

import (
	"log"

	"github.com/mokiat/gocrane/internal/project"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func printSummary(summary *project.Summary) {
	visited := maps.Keys(summary.Visited)
	slices.Sort(visited)
	errored := maps.Keys(summary.Errored)
	slices.Sort(errored)
	omitted := maps.Keys(summary.Omitted)
	slices.Sort(omitted)
	watchedDirs := maps.Keys(summary.WatchedDirs)
	slices.Sort(watchedDirs)
	watchedSourceFiles := maps.Keys(summary.WatchedSourceFiles)
	slices.Sort(watchedSourceFiles)
	watchedResourceFiles := maps.Keys(summary.WatchedResourceFiles)
	slices.Sort(watchedResourceFiles)

	log.Printf("Visited %d files or folders", len(visited))
	for _, file := range visited {
		log.Printf("\t Visited: %s", file)
	}

	log.Printf("Failed with %d files or folders", len(errored))
	for _, file := range errored {
		err := summary.Errored[file]
		log.Printf("\t Failure: %s (%s)", file, err)
	}

	log.Printf("Omitted %d files or folders", len(omitted))
	for _, file := range omitted {
		log.Printf("\t Omitted: %s", file)
	}

	log.Printf("Found %d directories to watch", len(watchedDirs))
	for _, dir := range watchedDirs {
		log.Printf("\t Watch dir: %s", dir)
	}

	log.Printf("Found %d source files (to use as digest)", len(watchedSourceFiles))
	for _, file := range watchedSourceFiles {
		log.Printf("\t Source file: %s", file)
	}

	log.Printf("Found %d resource files", len(watchedResourceFiles))
	for _, file := range watchedResourceFiles {
		log.Printf("\t Resource file: %s", file)
	}
}
