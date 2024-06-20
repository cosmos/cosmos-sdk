package schemagen

import (
	"pgregory.net/rapid"

	"cosmossdk.io/schema"
)

var Name = rapid.StringMatching(schema.NameFormat)
