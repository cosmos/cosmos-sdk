package v5_test

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	v5 "github.com/cosmos/cosmos-sdk/x/staking/migrations/v5"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
)

func TestMigrate(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(staking.AppModuleBasic{}).Codec
	storeKey := storetypes.NewKVStoreKey(v5.ModuleName)
	tKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)

	accAddrs := sims.CreateIncrementalAccounts(11)
	valAddrs := sims.ConvertAddrsToValAddrs(accAddrs[0:1])
	var addedDels []types.Delegation

	for i := 1; i < 11; i++ {
		del1 := types.NewDelegation(accAddrs[i], valAddrs[0], sdk.NewDec(100))
		store.Set(types.GetDelegationKey(accAddrs[i], valAddrs[0]), types.MustMarshalDelegation(cdc, del1))
		addedDels = append(addedDels, del1)
	}

	// before migration the state of delegations by val index should be empty
	dels := getValDelegations(ctx, cdc, storeKey, valAddrs[0])
	assert.Len(t, dels, 0)

	_, err := v5.MigrateStore(ctx, storeKey, cdc)
	assert.NoError(t, err)

	// after migration the state of delegations by val index should not be empty
	dels = getValDelegations(ctx, cdc, storeKey, valAddrs[0])
	assert.Len(t, dels, len(addedDels))
	assert.Equal(t, addedDels, dels)
}

func getValDelegations(ctx sdk.Context, cdc codec.Codec, storeKey storetypes.StoreKey, valAddr sdk.ValAddress) []types.Delegation {
	var delegations []types.Delegation

	store := ctx.KVStore(storeKey)
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
