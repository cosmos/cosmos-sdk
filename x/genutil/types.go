package genutil

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
)

type (
	// AppMap map modules names with their json raw representation
	AppMap map[string]json.RawMessage
	// MigrationCallback converts a genesis map from the previous version to the targeted one
	MigrationCallback func(AppMap, *codec.Codec) AppMap
	// MigrationMap defines a mapping from a version to a MigrationCallback
	MigrationMap map[string]MigrationCallback
)
