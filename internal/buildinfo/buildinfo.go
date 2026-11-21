// Package buildinfo carries build-time metadata injected with -ldflags.
package buildinfo

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

// ShortCommit returns the abbreviated commit hash used in the admin UI.
func ShortCommit() string {
	if len(Commit) > 7 {
		return Commit[:7]
	}
	return Commit
}
