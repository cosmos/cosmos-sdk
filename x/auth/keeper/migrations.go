package keeper

import (
	"github.com/cosmos/gogoproto/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v6 "github.com/cosmos/cosmos-sdk/x/auth/migrations/v6"
	v7 "github.com/cosmos/cosmos-sdk/x/auth/migrations/v7"
	v8 "github.com/cosmos/cosmos-sdk/x/auth/migrations/v8"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper      AccountKeeper
	queryServer grpc.Server
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper AccountKeeper, queryServer grpc.Server) Migrator {
	return Migrator{keeper: keeper, queryServer: queryServer}
}

// Migrate5to6 migrates the x/auth module state from the consensus version 5 to
// version 6. Specifically, it removes the global account number from storage.
func (m Migrator) Migrate5to6(ctx sdk.Context) error {
	return v6.Migrate(ctx, m.keeper.storeService, m.keeper.AccountNumber)
}

// Migrate6to7 migrates the x/auth module state from the consensus version 6 to
// version 7. Specifically, it updates the Params object.
func (m Migrator) Migrate6to7(ctx sdk.Context) error {
	return v7.Migrate(ctx, m.keeper.Params)
}

// Migrate7to8 migrates the x/auth module state from the consensus version 7 to
// version 8. Specifically, it adds the pubkey change params to the Params object.
func (m Migrator) Migrate7to8(ctx sdk.Context) error {
	return v8.Migrate(ctx, m.keeper.Params)
}
