package docs

import "embed"

//go:embed *.md
var Docs embed.FS
