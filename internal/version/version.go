package version

import "runtime/debug"

// Summary returns the application version as a string.
func Summary() string {
	buildInfo, _ := debug.ReadBuildInfo()

	if buildInfo.Main.Version != "" {
		return buildInfo.Main.Version
	}

	return "unknown"
}
