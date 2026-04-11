package command

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"maps"
	"slices"

	"github.com/mokiat/gocrane/internal/project"
)

func printSummary(summary *project.Summary) {
	visited := slices.Collect(maps.Keys(summary.Visited))
	slices.Sort(visited)
	errored := slices.Collect(maps.Keys(summary.Errored))
	slices.Sort(errored)
	omitted := slices.Collect(maps.Keys(summary.Omitted))
	slices.Sort(omitted)
	watchedDirs := slices.Collect(maps.Keys(summary.WatchedDirs))
	slices.Sort(watchedDirs)
	watchedSourceFiles := slices.Collect(maps.Keys(summary.WatchedSourceFiles))
	slices.Sort(watchedSourceFiles)
	watchedResourceFiles := slices.Collect(maps.Keys(summary.WatchedResourceFiles))
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

func calculateDigest(summary *project.Summary) (string, error) {
	sourceFiles := slices.Collect(maps.Keys(summary.WatchedSourceFiles))
	slices.Sort(sourceFiles) // ensure consistent order for digest calculation

	dig := sha256.New()
	for _, file := range sourceFiles {
		if err := project.WriteFileDigest(dig, string(file)); err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(dig.Sum(nil)), nil
}
