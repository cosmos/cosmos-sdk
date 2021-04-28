package module

import (
	"github.com/gogo/protobuf/grpc"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Configurator provides the hooks to allow modules to configure and register
// their services in the RegisterServices method. It is designed to eventually
// support module object capabilities isolation as described in
// https://github.com/cosmos/cosmos-sdk/issues/7093
type Configurator interface {
	// MsgServer returns a grpc.Server instance which allows registering services
	// that will handle TxBody.messages in transactions. These Msg's WILL NOT
	// be exposed as gRPC services.
	MsgServer() grpc.Server

	// QueryServer returns a grpc.Server instance which allows registering services
	// that will be exposed as gRPC services as well as ABCI query handlers.
	QueryServer() grpc.Server

	// RegisterMigration registers an in-place store migration for a module. The
	// handler is a migration script to perform in-place migrations from version
	// `forVersion` to version `forVersion+1`.
	//
	// EACH TIME a module's ConsensusVersion increments, a new migration MUST
	// be registered using this function. If a migration handler is missing for
	// a particular function, the upgrade logic (see RunMigrations function)
	// will panic. If the ConsensusVersion bump does not introduce any store
	// changes, then a no-op function must be registered here.
	RegisterMigration(moduleName string, forVersion uint64, handler MigrationHandler) error
}

type configurator struct {
	cdc         codec.Codec
	msgServer   grpc.Server
	queryServer grpc.Server

	// migrations is a map of moduleName -> forVersion -> migration script handler
	migrations map[string]map[uint64]MigrationHandler
}

// NewConfigurator returns a new Configurator instance
func NewConfigurator(cdc codec.Codec, msgServer grpc.Server, queryServer grpc.Server) Configurator {
	return configurator{
		cdc:         cdc,
		msgServer:   msgServer,
		queryServer: queryServer,
		migrations:  map[string]map[uint64]MigrationHandler{},
	}
}

var _ Configurator = configurator{}

// MsgServer implements the Configurator.MsgServer method
func (c configurator) MsgServer() grpc.Server {
	return c.msgServer
}

// QueryServer implements the Configurator.QueryServer method
func (c configurator) QueryServer() grpc.Server {
	return c.queryServer
}

// RegisterMigration implements the Configurator.RegisterMigration method
func (c configurator) RegisterMigration(moduleName string, forVersion uint64, handler MigrationHandler) error {
	if forVersion == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidVersion, "module migration versions should start at 1")
	}

	if c.migrations[moduleName] == nil {
		c.migrations[moduleName] = map[uint64]MigrationHandler{}
	}

	if c.migrations[moduleName][forVersion] != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrLogic, "another migration for module %s and version %d already exists", moduleName, forVersion)
	}

	c.migrations[moduleName][forVersion] = handler

	return nil
}

// runModuleMigrations runs all in-place store migrations for one given module from a
// version to another version.
func (c configurator) runModuleMigrations(ctx sdk.Context, moduleName string, fromVersion, toVersion uint64) error {
	// No-op if toVersion is the initial version.
	if toVersion <= 1 {
		return nil
	}

	moduleMigrationsMap, found := c.migrations[moduleName]
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "no migrations found for module %s", moduleName)
	}

	// Run in-place migrations for the module sequentially until toVersion.
	for i := fromVersion; i < toVersion; i++ {
		migrateFn, found := moduleMigrationsMap[i]
		if !found {
			return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "no migration found for module %s from version %d to version %d", moduleName, i, i+1)
		}

		err := migrateFn(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
