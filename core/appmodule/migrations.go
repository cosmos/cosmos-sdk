package appmodule

import "context"

type MigrationRegistrar interface {
	// Register registers an in-place store migration for a module. The
	// handler is a migration script to perform in-place migrations from version
	// `fromVersion` to version `fromVersion+1`.
	//
	// EACH TIME a module's ConsensusVersion increments, a new migration MUST
	// be registered using this function. If a migration handler is missing for
	// a particular function, the upgrade logic (see RunMigrations function)
	// will panic. If the ConsensusVersion bump does not introduce any store
	// changes, then a no-op function must be registered here.
	Register(moduleName string, fromVersion uint64, handler MigrationHandler) error
}

// MigrationHandler is the migration function that each module registers.
type MigrationHandler func(context.Context) error

// VersionMap is a map of moduleName -> version
type VersionMap map[string]uint64
