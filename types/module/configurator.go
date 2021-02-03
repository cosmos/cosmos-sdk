package module

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/grpc"
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
	// `fromVersion` to version `fromVersion+1`.
	RegisterMigration(moduleName string, fromVersion uint64, handler func(store sdk.KVStore) error) error
}

type configurator struct {
	msgServer   grpc.Server
	queryServer grpc.Server

	// storeKeys is used to access module stores inside in-place store migrations.
	storeKeys map[string]*sdk.KVStoreKey
	// migrations is a map of moduleName -> fromVersion -> migration script handler
	migrations map[string]map[uint64]func(store sdk.KVStore) error
}

// NewConfigurator returns a new Configurator instance
func NewConfigurator(msgServer grpc.Server, queryServer grpc.Server, storeKeys map[string]*sdk.KVStoreKey) Configurator {
	return configurator{
		msgServer:   msgServer,
		queryServer: queryServer,
		storeKeys:   storeKeys,
		migrations:  map[string]map[uint64]func(store sdk.KVStore) error{},
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
func (c configurator) RegisterMigration(moduleName string, fromVersion uint64, handler func(store sdk.KVStore) error) error {
	if c.migrations[moduleName][fromVersion] != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrLogic, "another migration for module %s and version %d already exists", moduleName, fromVersion)
	}

	_, found := c.storeKeys[moduleName]
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "store key for module %s not found", moduleName)
	}

	if c.migrations[moduleName] == nil {
		c.migrations[moduleName] = map[uint64]func(store sdk.KVStore) error{}
	}

	c.migrations[moduleName][fromVersion] = handler

	return nil
}
