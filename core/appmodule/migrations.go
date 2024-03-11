package appmodule

import (
	"cosmossdk.io/core/appmodule/v2"
)

// HasConsensusVersion is the interface for declaring a module consensus version.
type HasConsensusVersion = appmodule.HasConsensusVersion

// HasMigrations is the extension interface that modules should implement to register migrations.
type HasMigrations = appmodule.HasMigrations

type MigrationRegistrar = appmodule.MigrationRegistrar

// MigrationHandler is the migration function that each module registers.
type MigrationHandler = appmodule.MigrationHandler

// VersionMap is a map of moduleName -> version
type VersionMap = appmodule.VersionMap
