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
type MigrationMap map[string]func(from, to string) transform.Plan

var Migrations = MigrationMap{
	"v0.45": PlanBuilder,
	"v0.46": PlanBuilder,
	"v0.47": PlanBuilder,
}

// PlanBuilder is a function that returns a transformation plan for a given diff between two files.
func PlanBuilder(from, to string) transform.Plan {
	return nil
}
