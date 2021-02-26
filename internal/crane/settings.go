package crane

import "time"

type Settings struct {
	Verbose         bool
	Sources         []string
	Resources       []string
	Excludes        []string
	MainDir         string
	BinaryFile      string
	DigestFile      string
	BuildArgs       []string
	RunArgs         []string
	ShutdownTimeout time.Duration
}

type BuildSettings struct {
	Settings
}

type RunSettings struct {
	Settings
}
