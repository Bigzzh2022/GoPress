package migrations

import "embed"

// Files contains the versioned SQL migrations applied at service startup.
//
//go:embed *.up.sql
var Files embed.FS
