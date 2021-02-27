package command

import (
	"context"
	"fmt"
	"log"
	"os"

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
	log.Println("building binary...")
	builder := project.NewBuilder(cfg.MainDir, cfg.BuildArgs.Value(), cfg.BinaryFile)
	if err := builder.Build(ctx); err != nil {
		return fmt.Errorf("failed to build binary: %w", err)
	}
	log.Println("binary successfully built.")

	if cfg.DigestFile != "" {
		log.Println("analyzing project...")
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

		log.Println("calculating digest...")
		dig, err := layout.Digest()
		if err != nil {
			return fmt.Errorf("failed to calculate digest: %w", err)
		}
		log.Printf("digest: %s", dig)

		log.Println("persisting digest...")
		if err := writeDigest(cfg.DigestFile, dig); err != nil {
			return fmt.Errorf("failed to write digest: %w", err)
		}
		log.Println("digest successfully persisted.")
	}

	return nil
}

func writeDigest(file, digest string) error {
	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("failed to create file %q: %w", file, err)
	}
	defer f.Close()

	if _, err := f.WriteString(digest); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
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
	for _, file := range layout.ResourceFiles.SortedList() {
		log.Printf("resource file: %s", file)
	}

	log.Printf("found %d resource directories", len(layout.ResourceDirs))
	for _, dir := range layout.ResourceDirs.SortedList() {
		log.Printf("resource dir: %s", dir)
	}

	log.Printf("found %d source directories", len(layout.SourceDirs))
	for _, dir := range layout.SourceDirs.SortedList() {
		log.Printf("source dir: %s", dir)
	}

	log.Printf("found %d source files", len(layout.SourceFiles))
	for _, file := range layout.SourceFiles.SortedList() {
		log.Printf("source file: %s", file)
	}
}
