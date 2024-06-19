package indexerrapid

import (
	"pgregory.net/rapid"

	indexerbase "cosmossdk.io/indexer/base"
)

var Name = rapid.StringMatching(indexerbase.NameFormat)
