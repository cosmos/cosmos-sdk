package genutil

import (
	"encoding/json"
)

type (
	// AppMap map modules names with their json raw representation
	AppMap map[string]json.RawMessage
	// MigrationCallback converts a genesis map from the previous version to the targeted one
	MigrationCallback func(AppMap) AppMap
	// MigrationMap defines a mapping from a version to a MigrationCallback
	MigrationMap map[string]MigrationCallback
)
