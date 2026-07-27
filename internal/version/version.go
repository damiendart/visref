// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

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
