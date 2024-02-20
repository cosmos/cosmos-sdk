package appmodule

import (
	"cosmossdk.io/core/appmodule/v2"
)

type MigrationRegistrar = appmodule.MigrationRegistrar

// MigrationHandler is the migration function that each module registers.
type MigrationHandler = appmodule.MigrationHandler

// VersionMap is a map of moduleName -> version
type VersionMap = appmodule.VersionMap
