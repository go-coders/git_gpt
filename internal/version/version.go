package version

import (
	"fmt"
	"runtime"
)

var (
	// Version holds the current version of GitChat
	Version = "dev"

	// BuildTime will be injected during build
	BuildTime = "unknown"

	// GitCommit will be injected during build
	GitCommit = "unknown"
)

// GetVersionInfo returns a formatted version string
func GetVersionInfo() string {
	commit := GitCommit
	if len(commit) > 7 {
		commit = commit[:7]
	}

	if BuildTime == "unknown" && GitCommit == "unknown" {
		return fmt.Sprintf("%s (%s/%s)",
			Version,
			runtime.GOOS,
			runtime.GOARCH,
		)
	}

	return fmt.Sprintf("%s (%s/%s) - Build: %s, Commit: %s",
		Version,
		runtime.GOOS,
		runtime.GOARCH,
		BuildTime,
		commit,
	)
}

// GetVersion returns just the version number
func GetVersion() string {
	return Version
}
