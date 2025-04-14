package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"github.com/cosmos/gogoproto/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	v2 "github.com/cosmos/cosmos-sdk/x/auth/migrations/v2"
	v3 "github.com/cosmos/cosmos-sdk/x/auth/migrations/v3"
	v4 "github.com/cosmos/cosmos-sdk/x/auth/migrations/v4"
	v5 "github.com/cosmos/cosmos-sdk/x/auth/migrations/v5"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
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

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	var iterErr error

	m.keeper.IterateAccounts(ctx, func(account sdk.AccountI) (stop bool) {
		wb, err := v2.MigrateAccount(ctx, account, m.queryServer)
		if err != nil {
			iterErr = err
			return true
		}

		if wb == nil {
			return false
		}

		m.keeper.SetAccount(ctx, wb)
		return false
	})

	return iterErr
}

// Migrate2to3 migrates from consensus version 2 to version 3. Specifically, for each account
// we index the account's ID to their address.
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	return v3.MigrateStore(ctx, m.keeper.storeService, m.keeper.cdc)
}

// Migrate3to4 migrates the x/auth module state from the consensus version 3 to
// version 4. Specifically, it takes the parameters that are currently stored
// and managed by the x/params modules and stores them directly into the x/auth
// module state.
func (m Migrator) Migrate3to4(ctx sdk.Context) error {
	return v4.Migrate(ctx, m.keeper.storeService, m.legacySubspace, m.keeper.cdc)
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

// AddAddressSpace adds all the address mappings for a newly defined address space.
// This migration must be run whenever a new address space is defined with AccountKeeper.DefineAddressSpace.
func (m Migrator) AddAddressSpace(ctx sdk.Context, addressSpace string) error {
	// iterate over all existing accounts and add their address space mappings
	acctIt, err := m.keeper.Accounts.Iterate(ctx, nil)
	if err != nil {
		return err
	}

	addrPrefix, ok := m.keeper.addressPrefixByName[addressSpace]
	if !ok {
		return fmt.Errorf("address space %s not found", addressSpace)
	}

	addrMgr, ok := m.keeper.addressSpaceManagers[addrPrefix]
	if !ok {
		return fmt.Errorf("address space manager %s not found", addrPrefix)
	}

	for acctIt.Valid() {
		acct, err := acctIt.Value()
		if err != nil {
			return err
		}
		num := acct.GetAccountNumber()
		id := accountNumToId(num)
		pk := acct.GetPubKey()
		newAddr := addrMgr.DeriveAddress(id, pk)
		if newAddr == nil {
			return fmt.Errorf("failed to derive address for account %s", acct.GetAddress())
		}

		err = m.keeper.AddressByAccountID.Set(ctx, collections.Join(id, addrPrefix), newAddr)
		if err != nil {
			return err
		}
		err = m.keeper.AccountIDByAddress.Set(ctx, collections.Join(addrPrefix, newAddr), id)
		if err != nil {
			return err
		}

		acctIt.Next()
	}
	panic("TODO")
}
