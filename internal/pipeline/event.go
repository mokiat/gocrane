package pipeline

type ChangeEvent struct {
	Paths []string
}

// InvalidationEvent indicates that the application needs to be restarted and
// optionally rebuilt as well.
type InvalidationEvent struct {

	// ShouldRebuild indicates whether the application should be rebuilt before
	// being restarted. If false, the application will be restarted with the
	// currently available binary. If true, the application will be rebuilt before
	// being restarted.
	ShouldRebuild bool
}
