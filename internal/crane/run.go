package crane

// func Run(ctx context.Context, settings Settings) error {
// 	// Trigger an initial build by faking a change if we don't have
// 	// a cached executable to use.
// 	if settings.CachedBuild == "" {
// 		go func() {
// 			changeEventQueue <- events.Change{}
// 		}()
// 	}

// 	// Trigger an initial run by faking a build change if we have
// 	// a cached executable specified.
// 	if settings.CachedBuild != "" {
// 		go func() {
// 			buildEventQueue <- events.Build{
// 				Path: settings.CachedBuild,
// 			}
// 		}()
// 	}

// }
