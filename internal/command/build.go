package command

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/urfave/cli/v2"

	"github.com/mokiat/gocrane/internal/flag"
	"github.com/mokiat/gocrane/internal/project"
)

func Build() *cli.Command {
	var cfg buildConfig
	return &cli.Command{
		Name: "build",
		Flags: []cli.Flag{
			newVerboseFlag(&cfg.Verbose),
			newSourcesFlag(&cfg.Sources),
			newResourcesFlag(&cfg.Resources),
			newExcludesFlag(&cfg.Excludes),
			newMainFlag(&cfg.MainDir),
			newBinaryFlag(&cfg.BinaryFile, true),
			newDigestFlag(&cfg.DigestFile),
			newBuildArgs(&cfg.BuildArgs),
		},
		Action: func(c *cli.Context) error {
			return build(c.Context, cfg)
		},
	}
}

type buildConfig struct {
	Verbose    bool
	Sources    cli.StringSlice
	Resources  cli.StringSlice
	Excludes   cli.StringSlice
	MainDir    string
	BinaryFile string
	DigestFile string
	BuildArgs  flag.ShlexStringSlice
}

func build(ctx context.Context, cfg buildConfig) error {
	log.Println("building application...")

	layout, err := project.Explore(
		cfg.Sources.Value(),
		cfg.Resources.Value(),
		cfg.Excludes.Value(),
	)
	if err != nil {
		return fmt.Errorf("failed to explore project: %w", err)
	}
	if cfg.Verbose {
		logLayout(layout)
	}
	return nil
}

func logLayout(layout *project.Layout) {
	log.Printf("omitted %d files or folders", len(layout.Omitted))
	for file, err := range layout.Omitted {
		if err != nil {
			log.Printf("omitted: %s; reason: %s", file, err)
		} else {
			log.Printf("omitted: %s", file)
		}
	}

	log.Printf("found %d resource files", len(layout.ResourceFiles))
	for _, file := range fileSetToSortedSlice(layout.ResourceFiles) {
		log.Printf("resource file: %s", file)
	}

	log.Printf("found %d resource directories", len(layout.ResourceDirs))
	for _, dir := range fileSetToSortedSlice(layout.ResourceDirs) {
		log.Printf("resource dir: %s", dir)
	}

	log.Printf("found %d source directories", len(layout.SourceDirs))
	for _, dir := range fileSetToSortedSlice(layout.SourceDirs) {
		log.Printf("source dir: %s", dir)
	}

	log.Printf("found %d source files", len(layout.SourceFiles))
	for _, file := range fileSetToSortedSlice(layout.SourceFiles) {
		log.Printf("source file: %s", file)
	}
}

func fileSetToSortedSlice(files map[string]struct{}) []string {
	result := make([]string, 0, len(files))
	for file := range files {
		result = append(result, file)
	}
	sort.Strings(result)
	return result
}
