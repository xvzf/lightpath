package version

import "fmt"

var (
	commit = ""
	date   = ""
	tag    = ""
)

// GetCommit returns the current commit.
func GetCommit() string {
	return commit
}

// GetTag returns the current commit.
func GetTag() string {
	return tag
}

// GetBuildDate returns the build date.
func GetBuildDate() string {
	return date
}

// GetVersion returns a version string.
func GetVersion() string {
	return fmt.Sprintf("tag=%s commit=%s date=%s", GetTag(), GetCommit(), GetBuildDate())
}
