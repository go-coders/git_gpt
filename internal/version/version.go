package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
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

func InitFromBuildInfo(info *debug.BuildInfo) {
	// Set version from main module if available
	if info.Main.Version != "(devel)" {
		Version = strings.TrimPrefix(info.Main.Version, "v")
	}

	// Get version from dependencies if not found in main
	for _, dep := range info.Deps {
		if dep.Path == "github.com/go-coders/git_gpt" {
			Version = strings.TrimPrefix(dep.Version, "v")
			break
		}
	}

	// Get vcs information from settings
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			GitCommit = setting.Value
		case "vcs.time":
			BuildTime = setting.Value
		}
	}
}
