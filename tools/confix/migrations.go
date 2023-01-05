package confix

import (
	"github.com/creachadair/tomledit/transform"
)

const (
	AppConfig    = "app.toml"
	ClientConfig = "client.toml"
	TMConfig     = "config.toml"
)

// MigrationMap defines a mapping from a version to a transformation plan.
type MigrationMap map[string]transform.Plan

var Migrations = MigrationMap{
	"v0.45": nil,
	"v0.46": nil,
	"v0.47": nil,
	"next":  nil, // unreleased version of the SDK
}
