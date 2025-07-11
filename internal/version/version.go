package version

// These variables will be set at build time using -ldflags
var (
	// Version is the current version of the CLI
	Version = "dev"

	// CommitHash is the git commit hash of the build
	CommitHash = "unknown"

	// BuildDate is the date when the binary was built
	BuildDate = "unknown"

	// DisableAutoUpdate is a flag to disable auto-update functionality
	DisableAutoUpdate = "false"
)

// GetVersion returns the current version of the CLI
func GetVersion() string {
	return Version
}

// GetCommitHash returns the git commit hash of the build
func GetCommitHash() string {
	return CommitHash
}

// GetBuildDate returns the date when the binary was built
func GetBuildDate() string {
	return BuildDate
}

// IsAutoUpdateDisabled returns true if auto-update functionality is disabled
func IsAutoUpdateDisabled() bool {
	return DisableAutoUpdate == "true"
}
