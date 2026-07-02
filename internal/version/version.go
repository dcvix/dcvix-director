//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package version

import "fmt"

var (
	// Vars set up at build time via "-ldflags"
	// Version holds the current version injected at build time from git tag
	Version = "dev"
	// Commit holds the git commit hash
	Commit = "none"
	// BuildTime holds the build timestamp
	BuildTime = "unknown"
)

// String returns a formatted version string
func String() string {
	return fmt.Sprintf("%s (commit: %s, built at: %s)", Version, Commit, BuildTime)
}
