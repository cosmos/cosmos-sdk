package keeper

import (
	"github.com/cosmos/gogoproto/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v5 "github.com/cosmos/cosmos-sdk/x/auth/migrations/v5"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
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

// Migrate4To5 migrates the x/auth module state from the consensus version 4 to 5.
// It migrates the GlobalAccountNumber from being a protobuf defined value to a
// big-endian encoded uint64, it also migrates it to use a more canonical prefix.
func (m Migrator) Migrate4To5(ctx sdk.Context) error {
	return v5.Migrate(ctx, m.keeper.storeService, m.keeper.AccountNumber)
}

// V45_SetAccount implements V45_SetAccount
// set the account without map to accAddr to accNumber.
//
// NOTE: This is used for testing purposes only.
func (m Migrator) V45SetAccount(ctx sdk.Context, acc sdk.AccountI) error {
	addr := acc.GetAddress()
	store := m.keeper.storeService.OpenKVStore(ctx)

	bz, err := m.keeper.Accounts.ValueCodec().Encode(acc)
	if err != nil {
		return err
	}

	return store.Set(addressStoreKey(addr), bz)
}

// addressStoreKey turn an address to key used to get it from the account store
// NOTE(tip): exists for legacy compatibility
func addressStoreKey(addr sdk.AccAddress) []byte {
	return append(types.AddressStoreKeyPrefix, addr.Bytes()...)
}
