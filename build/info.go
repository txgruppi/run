package build

import "time"

var (
	// Version is the current version of the application
	Version = "0.0.0"
	// Commit is the git commit hash of the current version
	Commit = ""
	// Compiled is the time the application was compiled
	Compiled string
)

// CompiledTime return a time.Time instance with the parsed value of Compiled
func CompiledTime() time.Time {
	t, _ := time.Parse("2006-01-02T15:04:05Z-0700", Compiled)
	return t
}
