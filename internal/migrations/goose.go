package migrations

import "embed"

//go:embed sql/schema/*.sql
var GooseFS embed.FS
