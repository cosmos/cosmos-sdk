package schematesting

import (
	"pgregory.net/rapid"

	"cosmossdk.io/schema"
)

var NameGen = rapid.StringMatching(schema.NameFormat)
