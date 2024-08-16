package schematesting

import (
	"pgregory.net/rapid"

	"cosmossdk.io/schema"
)

// NameGen validates valid names that match the NameFormat regex.
var NameGen = rapid.StringMatching(schema.NameFormat)
