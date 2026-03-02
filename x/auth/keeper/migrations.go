package keeper

import (
	"github.com/cosmos/gogoproto/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	v6 "github.com/cosmos/cosmos-sdk/x/auth/migrations/v6"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper         AccountKeeper
	queryServer    grpc.Server
	legacySubspace exported.Subspace
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper AccountKeeper, queryServer grpc.Server, ss exported.Subspace) Migrator {
	return Migrator{keeper: keeper, queryServer: queryServer, legacySubspace: ss}
}

// Migrate5to6 migrates the x/auth module state from the consensus version 5 to
// version 6. Specifically, it removes the global account number from storage.
func (m Migrator) Migrate5to6(ctx sdk.Context) error {
	return v6.Migrate(ctx, m.keeper.storeService, m.keeper.AccountNumber)
}
