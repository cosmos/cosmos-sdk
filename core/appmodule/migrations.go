package appmodule

import (
	"cosmossdk.io/core/appmodule/v2"
)

// HasConsensusVersion is the interface for declaring a module consensus version.
type HasConsensusVersion = appmodule.HasConsensusVersion

// HasMigrations is implemented by a module which upgrades or has upgraded to a new consensus version.
type HasMigrations = appmodule.HasMigrations

// MigrationRegistrar is the interface for registering in-place store migrations.
type MigrationRegistrar = appmodule.MigrationRegistrar

// MigrationHandler is the migration function that each module registers.
type MigrationHandler = appmodule.MigrationHandler

// VersionMap is a map of moduleName -> version
type VersionMap = appmodule.VersionMap
