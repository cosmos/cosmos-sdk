package runtime

import (
	"context"
	"fmt"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
)

var _ appmodulev2.MigrationRegistrar = (*migrationRegistrar)(nil)

type migrationRegistrar struct {
	// migrations is a map of moduleName -> fromVersion -> migration script handler
	migrations map[string]map[uint64]appmodulev2.MigrationHandler
}

// newMigrationRegistrar is a constructor for registering in-place store migrations for modules.
func newMigrationRegistrar() *migrationRegistrar {
	return &migrationRegistrar{
		migrations: make(map[string]map[uint64]appmodulev2.MigrationHandler),
	}
}

// Register registers an in-place store migration for a module.
// It permits to register modules migrations that have migrated to serverv2 but still be compatible with baseapp.
func (mr *migrationRegistrar) Register(moduleName string, fromVersion uint64, handler appmodulev2.MigrationHandler) error {
	if fromVersion == 0 {
		return fmt.Errorf("module migration versions should start at 1: %s", moduleName)
	}

	if mr.migrations[moduleName] == nil {
		mr.migrations[moduleName] = map[uint64]appmodulev2.MigrationHandler{}
	}

	if mr.migrations[moduleName][fromVersion] != nil {
		return fmt.Errorf("another migration for module %s and version %d already exists", moduleName, fromVersion)
	}

	mr.migrations[moduleName][fromVersion] = handler

	return nil
}

// RunModuleMigrations runs all in-place store migrations for one given module from a version to another version.
func (mr *migrationRegistrar) RunModuleMigrations(ctx context.Context, moduleName string, fromVersion, toVersion uint64) error {
	// No-op if toVersion is the initial version or if the version is unchanged.
	if toVersion <= 1 || fromVersion == toVersion {
		return nil
	}

	moduleMigrationsMap, found := mr.migrations[moduleName]
	if !found {
		return fmt.Errorf("no migrations found for module %s", moduleName)
	}

	// Run in-place migrations for the module sequentially until toVersion.
	for i := fromVersion; i < toVersion; i++ {
		migrateFn, found := moduleMigrationsMap[i]
		if !found {
			return fmt.Errorf("no migration found for module %s from version %d to version %d", moduleName, i, i+1)
		}

		if err := migrateFn(ctx); err != nil {
			return err
		}
	}

	return nil
}
