package pipeline

// Deprecated: Switch to ChangeEvent and new node mechanism.
type OldChangeEvent struct {
	Paths []string
}

// ChangeEvent represents a change to a watched file or folder.
type ChangeEvent struct {

	// Path is the absolute path to the file or folder that has changed.
	Path string
}

// RestartEvent indicates that the application needs to be restarted and
// optionally rebuilt as well.
type RestartEvent struct {

	// ShouldRebuild indicates whether the application should be rebuilt before
	// being restarted. If false, the application will be restarted with the
	// currently available binary. If true, the application will be rebuilt before
	// being restarted.
	ShouldRebuild bool
}
