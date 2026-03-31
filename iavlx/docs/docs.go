package docs

import "embed"

// Docs embeds the markdown documentation files for the iavlx storage engine.
// Used by the iavlx CLI's help system.
//
//go:embed *.md
var Docs embed.FS
