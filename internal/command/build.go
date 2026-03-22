package command

import (
	"context"
	"fmt"
	"log"

	"github.com/urfave/cli/v3"

	"github.com/mokiat/gocrane/internal/command/flag"
	"github.com/mokiat/gocrane/internal/project"
)

func Build() *cli.Command {
	var cfg buildConfig
	return &cli.Command{
		Name:        "build",
		Usage:       "build a cached binary of the go application",
		Description: "use this command during image build to prepare a cached version of the application binary for faster startup",
		Flags: []cli.Flag{
			newVerboseFlag(&cfg.Verbose),
			newDirFlag(&cfg.Dirs),
			newDirExcludeFlag(&cfg.ExcludeDirs),
			newSourceFlag(&cfg.Sources),
			newSourceExcludeFlag(&cfg.ExcludeSources),
			newResourceFlag(&cfg.Resources),
			newResourceExcludeFlag(&cfg.ExcludeResources),
			newMainFlag(&cfg.MainDir),
			newBinaryFlag(&cfg.BinaryFile, true),
			newBuildArgs(&cfg.BuildArgs),
		},
		Action: func(ctx context.Context, _ *cli.Command) error {
			return build(ctx, cfg)
		},
	}
}

type buildConfig struct {
	Verbose          bool
	Dirs             []string
	ExcludeDirs      []string
	Sources          []string
	ExcludeSources   []string
	Resources        []string
	ExcludeResources []string
	MainDir          string
	BinaryFile       string
	BuildArgs        flag.ShlexStringSlice
}

func build(ctx context.Context, cfg buildConfig) error {
	log.Println("Building binary...")
	builder := project.NewBuilder(cfg.MainDir, cfg.BuildArgs.Items())
	if err := builder.Build(ctx, cfg.BinaryFile); err != nil {
		return fmt.Errorf("failed to build binary: %w", err)
	}

	log.Println("Preparing filtering...")
	watchFilter, err := buildFilterTree(cfg.Dirs, cfg.ExcludeDirs)
	if err != nil {
		return fmt.Errorf("problem with dir rules: %w", err)
	}
	sourceFilter, err := buildFilterTree(cfg.Sources, cfg.ExcludeSources)
	if err != nil {
		return fmt.Errorf("problem with source rules: %w", err)
	}
	resourceFilter, err := buildFilterTree(cfg.Resources, cfg.ExcludeResources)
	if err != nil {
		return fmt.Errorf("problem with resource rules: %w", err)
	}
	rootDirs := watchFilter.RootPaths()

	log.Println("Analyzing project...")
	summary := project.Analyze(rootDirs, watchFilter, sourceFilter, resourceFilter)
	if cfg.Verbose {
		printSummary(summary)
	}

	log.Println("Calculating current digest...")
	digest, err := calculateDigest(summary)
	if err != nil {
		return fmt.Errorf("failed to calculate digest: %w", err)
	}
	log.Printf("Digest: %s", digest)

	log.Println("Persisting digest...")
	digestFile := fmt.Sprintf("%s.dig", cfg.BinaryFile)
	if err := project.SaveDigestFile(digestFile, digest); err != nil {
		return fmt.Errorf("failed to write digest: %w", err)
	}
	log.Println("Digest successfully persisted.")

	log.Println("Done.")
	return nil
}
