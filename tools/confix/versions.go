package confix

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/creachadair/tomledit"
)

const (
	AppConfig    = "app.toml"
	ClientConfig = "client.toml"
	TMConfig     = "config.toml"
)

type (
	// MigrationCallback converts a config from the previous version to the targeted one.
	MigrationCallback func(tomledit.Document, client.Context) (tomledit.Document, error)

	// MigrationMap defines a mapping from a version to a MigrationCallback.
	MigrationMap map[string]MigrationCallback
)

var Versions = MigrationMap{
	"v0.45": nil,
	"v0.46": nil,
	"v0.47": nil,
	"next":  nil, // unreleased version of the SDK
}
