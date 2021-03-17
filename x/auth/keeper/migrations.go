package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/gogo/protobuf/grpc"
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

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	m.keeper.IterateAccounts(ctx, func(account types.AccountI) (stop bool) {
		return false
	})
	return nil
}
