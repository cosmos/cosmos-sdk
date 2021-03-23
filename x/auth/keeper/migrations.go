package keeper

import (
	v043 "github.com/cosmos/cosmos-sdk/x/auth/legacy/v043"
	"github.com/gogo/protobuf/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	wb, err := v043.MigrateStore(ctx, m.keeper.GetAllAccounts(ctx), m.queryServer)
	if err != nil {
		return err
	}

	for _, a := range wb {
		m.keeper.SetAccount(ctx, a)
	}

	return nil
}
