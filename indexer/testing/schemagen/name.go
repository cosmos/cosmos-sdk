package schemagen

import (
	indexerbase "cosmossdk.io/schema"
	"pgregory.net/rapid"
)

var Name = rapid.StringMatching(schema.NameFormat)
