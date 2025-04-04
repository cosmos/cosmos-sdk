package keeper_test

import (
	"cosmossdk.io/core/store"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TestDelegationsByValidatorMigration tests the multi block migration of the reverse delegation index
func (s *KeeperTestSuite) TestDelegationsByValidatorMigration() {
	require := s.Require()
	ctx, keeper := s.ctx, s.stakingKeeper
	store := s.storeService.OpenKVStore(ctx)
	storeInit := runtime.KVStoreAdapter(store)
	cdc := moduletestutil.MakeTestEncodingConfig(staking.AppModuleBasic{}).Codec

	accAddrs := sims.CreateIncrementalAccounts(15)
	valAddrs := sims.ConvertAddrsToValAddrs(accAddrs[0:1])
	var addedDels []types.Delegation

	// start at 1 as 0 addr is the validator addr
	for i := 1; i < 11; i++ {
		del1 := types.NewDelegation(accAddrs[i].String(), valAddrs[0].String(), sdkmath.LegacyNewDec(100))
		store.Set(types.GetDelegationKey(accAddrs[i], valAddrs[0]), types.MustMarshalDelegation(cdc, del1))
		addedDels = append(addedDels, del1)
	}

	// number of items we migrate per migration
	migrationCadence := 6

	// set the key in the store, this happens on the original migration
	iterator := storetypes.KVStorePrefixIterator(storeInit, types.DelegationKey)
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		store.Set(types.NextMigrateDelegationsByValidatorIndexKey, key[1:])
		break
	}

	// before migration the state of delegations by val index should be empty
	dels := getValDelegations(cdc, store, valAddrs[0])
	require.Equal(len(dels), 0)

	// run the first round of migrations first 6, 10 in store
	err := keeper.MigrateDelegationsByValidatorIndex(ctx, migrationCadence)
	require.NoError(err)

	// after migration the state of delegations by val index should not be empty
	dels = getValDelegations(cdc, store, valAddrs[0])
	require.Equal(len(dels), migrationCadence)
	require.NotEqual(len(dels), len(addedDels))

	// check that the next value needed from the store is present
	next, err := store.Get(types.NextMigrateDelegationsByValidatorIndexKey)
	require.NoError(err)
	require.NotNil(next)

	// delegate to a validator while the migration is in progress
	delagationWhileMigrationInProgress := types.NewDelegation(accAddrs[12].String(), valAddrs[0].String(), sdkmath.LegacyNewDec(100))
	keeper.SetDelegation(ctx, delagationWhileMigrationInProgress)
	addedDels = append(addedDels, delagationWhileMigrationInProgress)

	// remove a delegation from a validator while the migration is in progress that has been processed
	removeDelagationWhileMigrationInProgress := types.NewDelegation(accAddrs[3].String(), valAddrs[0].String(), sdkmath.LegacyNewDec(100))
	keeper.RemoveDelegation(ctx, removeDelagationWhileMigrationInProgress)
	// index in the array is 2
	addedDels = deleteElement(addedDels, 2)

	// remove the index on the off chance this happens during the migration
	removeDelagationWhileMigrationInProgressNextIndex := types.NewDelegation(accAddrs[6].String(), valAddrs[0].String(), sdkmath.LegacyNewDec(100))
	keeper.RemoveDelegation(ctx, removeDelagationWhileMigrationInProgressNextIndex)
	// index in the array is 4, as we've removed one item
	addedDels = deleteElement(addedDels, 4)

	// remove a delegation from a validator while the migration is in progress that has not been processed
	removeDelagationWhileMigrationInProgressNotProcessed := types.NewDelegation(accAddrs[10].String(), valAddrs[0].String(), sdkmath.LegacyNewDec(100))
	keeper.RemoveDelegation(ctx, removeDelagationWhileMigrationInProgressNotProcessed)
	// index in the array is 7, as we've removed 2 items
	addedDels = deleteElement(addedDels, 7)

	// while migrating get state of delegations by val index should be increased by 1
	delagationWhileMigrationInProgressCount := getValDelegations(cdc, store, valAddrs[0])
	require.Equal(len(delagationWhileMigrationInProgressCount), migrationCadence-1)

	// run the second round of migrations
	err = keeper.MigrateDelegationsByValidatorIndex(ctx, migrationCadence)
	require.NoError(err)

	// after migration the state of delegations by val index equal all delegations
	dels = getValDelegations(cdc, store, valAddrs[0])
	require.Equal(len(dels), len(addedDels))
	require.Equal(dels, addedDels)

	// check that the next value needed from the store is empty
	next, err = store.Get(types.NextMigrateDelegationsByValidatorIndexKey)
	require.NoError(err)
	require.Nil(next)

	// Iterate over the store by delegation key
	delKeyCount := 0
	iteratorDel := storetypes.KVStorePrefixIterator(storeInit, types.DelegationKey)
	for ; iteratorDel.Valid(); iteratorDel.Next() {
		delKeyCount++
	}

	// Iterate over the store by validator key
	valKeyCount := 0
	iteratorVal := storetypes.KVStorePrefixIterator(storeInit, types.DelegationByValIndexKey)
	for ; iteratorVal.Valid(); iteratorVal.Next() {
		valKeyCount++
	}

	// Make sure the store count is the same
	require.Equal(valKeyCount, delKeyCount)
}

// deleteElement is a simple helper function to remove items from a slice
func deleteElement(slice []types.Delegation, index int) []types.Delegation {
	return append(slice[:index], slice[index+1:]...)
}

// getValidatorDelegations is a helper function to get all delegations using the new v5 staking reverse index
func getValDelegations(cdc codec.Codec, keeperStore store.KVStore, valAddr sdk.ValAddress) []types.Delegation {
	var delegations []types.Delegation

	store := runtime.KVStoreAdapter(keeperStore)
	iterator := storetypes.KVStorePrefixIterator(store, types.GetDelegationsByValPrefixKey(valAddr))

	for ; iterator.Valid(); iterator.Next() {
		var delegation types.Delegation
		valAddr, delAddr, err := types.ParseDelegationsByValKey(iterator.Key())
		if err != nil {
			panic(err)
		}

		bz := store.Get(types.GetDelegationKey(delAddr, valAddr))
		cdc.MustUnmarshal(bz, &delegation)

		delegations = append(delegations, delegation)
	}

	return delegations
}
