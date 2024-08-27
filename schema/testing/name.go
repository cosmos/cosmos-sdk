package schematesting

import (
	"pgregory.net/rapid"

	"cosmossdk.io/schema"
)

// NameGen validates valid names that match the NameFormat regex.
var NameGen = rapid.StringMatching(schema.NameFormat)

// QualifiedNameGen validates valid qualified names that match the QualifiedNameFormat regex.
var QualifiedNameGen = rapid.StringMatching(schema.QualifiedNameFormat).Filter(func(x string) bool {
	return len(x) < 64
})
