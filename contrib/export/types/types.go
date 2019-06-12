package types

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/codec"
)

type (
	AppMap            map[string]json.RawMessage
	MigrationCallback func(AppMap, *codec.Codec) AppMap
	MigrationMap      map[string]MigrationCallback // It maps a version to a function migrate to the previous one to this.
	// We can expand this type to include the previous version too
)
