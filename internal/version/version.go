package version

import (
	"fmt"
	"runtime"
	"time"
)

// Version information - these will be set at build time via ldflags
var (
	// Version is the semantic version of the application
	Version = "dev"

	// GitCommit is the git commit hash when the binary was built
	GitCommit = "unknown"

	// GitTag is the git tag when the binary was built
	GitTag = "unknown"

	// BuildDate is when the binary was built
	BuildDate = "unknown"

	// BuildUser is who built the binary
	BuildUser = "unknown"
)

// Info contains detailed version information
type Info struct {
	Version   string    `json:"version"`
	GitCommit string    `json:"gitCommit"`
	GitTag    string    `json:"gitTag"`
	BuildDate string    `json:"buildDate"`
	BuildUser string    `json:"buildUser"`
	GoVersion string    `json:"goVersion"`
	Platform  string    `json:"platform"`
	Arch      string    `json:"arch"`
	Timestamp time.Time `json:"timestamp"`
}

// Get returns detailed version information
func Get() Info {
	buildTime, _ := time.Parse(time.RFC3339, BuildDate)
	if BuildDate == "unknown" {
		buildTime = time.Now()
	}

	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		GitTag:    GitTag,
		BuildDate: BuildDate,
		BuildUser: BuildUser,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS,
		Arch:      runtime.GOARCH,
		Timestamp: buildTime,
	}
}

// shortCommit returns the first 8 characters of a commit hash, or the full string if shorter
func shortCommit(commit string) string {
	if len(commit) > 8 {
		return commit[:8]
	}
	return commit
}

// String returns a formatted version string
func String() string {
	info := Get()

	if GitTag != "unknown" && GitTag != "" {
		return fmt.Sprintf("SysMind %s (%s)", GitTag, shortCommit(info.GitCommit))
	}

	if Version != "dev" {
		return fmt.Sprintf("SysMind v%s (%s)", Version, shortCommit(info.GitCommit))
	}

	return fmt.Sprintf("SysMind dev (%s)", shortCommit(info.GitCommit))
}

// Short returns a short version string
func Short() string {
	if GitTag != "unknown" && GitTag != "" {
		return GitTag
	}

	if Version != "dev" {
		return fmt.Sprintf("v%s", Version)
	}

	return "dev"
}

// UserAgent returns a user agent string for HTTP requests
func UserAgent() string {
	info := Get()
	return fmt.Sprintf("SysMind/%s (%s; %s)", Short(), info.Platform, info.Arch)
}

// PrintBuildInfo prints detailed build information to console
func PrintBuildInfo() {
	info := Get()
	fmt.Printf("SysMind Version Information:\n")
	fmt.Printf("  Version:    %s\n", info.Version)
	fmt.Printf("  Git Commit: %s\n", info.GitCommit)
	fmt.Printf("  Git Tag:    %s\n", info.GitTag)
	fmt.Printf("  Build Date: %s\n", info.BuildDate)
	fmt.Printf("  Build User: %s\n", info.BuildUser)
	fmt.Printf("  Go Version: %s\n", info.GoVersion)
	fmt.Printf("  Platform:   %s/%s\n", info.Platform, info.Arch)
}
