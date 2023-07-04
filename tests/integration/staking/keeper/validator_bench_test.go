package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func BenchmarkGetValidator(b *testing.B) {
	// 900 is the max number we are allowed to use in order to avoid simtestutil.CreateTestPubKeys
	// panic: encoding/hex: odd length hex string
	powersNumber := 900

	var totalPower int64
	powers := make([]int64, powersNumber)
	for i := range powers {
		powers[i] = int64(i)
		totalPower += int64(i)
	}

	f, _, valAddrs, vals := initValidators(b, totalPower, len(powers), powers)

	for _, validator := range vals {
		f.stakingKeeper.SetValidator(f.sdkCtx, validator)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, addr := range valAddrs {
			_, _ = f.stakingKeeper.GetValidator(f.sdkCtx, addr)
		}
	}
}

func BenchmarkGetValidatorDelegations(b *testing.B) {
	var totalPower int64
	powersNumber := 10

	powers := make([]int64, powersNumber)
	for i := range powers {
		powers[i] = int64(i)
		totalPower += int64(i)
	}

	f, _, valAddrs, vals := initValidators(b, totalPower, len(powers), powers)
	for _, validator := range vals {
		f.stakingKeeper.SetValidator(f.sdkCtx, validator)
	}

	delegationsNum := 1000
	for _, val := range valAddrs {
		for i := 0; i < delegationsNum; i++ {
			delegator := sdk.AccAddress(fmt.Sprintf("address%d", i))
			banktestutil.FundAccount(f.sdkCtx, f.bankKeeper, delegator,
				sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(int64(i)))))
			NewDel := types.NewDelegation(delegator, val, math.LegacyNewDec(int64(i)))
			f.stakingKeeper.SetDelegation(f.sdkCtx, NewDel)
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		updateValidatorDelegations(f, valAddrs[0], sdk.ValAddress("val"))
	}
}

func BenchmarkGetValidatorDelegationsLegacy(b *testing.B) {
	var totalPower int64
	powersNumber := 10

	powers := make([]int64, powersNumber)
	for i := range powers {
		powers[i] = int64(i)
		totalPower += int64(i)
	}

	f, _, valAddrs, vals := initValidators(b, totalPower, len(powers), powers)

	for _, validator := range vals {
		f.stakingKeeper.SetValidator(f.sdkCtx, validator)
	}

	delegationsNum := 1000
	for _, val := range valAddrs {
		for i := 0; i < delegationsNum; i++ {
			delegator := sdk.AccAddress(fmt.Sprintf("address%d", i))
			banktestutil.FundAccount(f.sdkCtx, f.bankKeeper, delegator,
				sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(int64(i)))))
			NewDel := types.NewDelegation(delegator, val, math.LegacyNewDec(int64(i)))
			f.stakingKeeper.SetDelegation(f.sdkCtx, NewDel)
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		updateValidatorDelegationsLegacy(f, valAddrs[0], sdk.ValAddress("val"))
	}
}

func updateValidatorDelegationsLegacy(f *fixture, existingValAddr, newValAddr sdk.ValAddress) {
	storeKey := f.keys[types.StoreKey]
	cdc, k := f.cdc, f.stakingKeeper

	store := f.sdkCtx.KVStore(storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.DelegationKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		delegation := types.MustUnmarshalDelegation(cdc, iterator.Value())
		if delegation.GetValidatorAddr().Equals(existingValAddr) {
			k.RemoveDelegation(f.sdkCtx, delegation)
			delegation.ValidatorAddress = newValAddr.String()
			k.SetDelegation(f.sdkCtx, delegation)
		}
	}
}

func updateValidatorDelegations(f *fixture, existingValAddr, newValAddr sdk.ValAddress) {
	storeKey := f.keys[types.StoreKey]
	cdc, k := f.cdc, f.stakingKeeper

	store := f.sdkCtx.KVStore(storeKey)

	itr := storetypes.KVStorePrefixIterator(store, types.GetDelegationsByValPrefixKey(existingValAddr))
	defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		key := itr.Key()
		valAddr, delAddr, err := types.ParseDelegationsByValKey(key)
		if err != nil {
			panic(err)
		}

		bz := store.Get(types.GetDelegationKey(delAddr, valAddr))
		delegation := types.MustUnmarshalDelegation(cdc, bz)

		// remove old operator addr from delegation
		if err := k.RemoveDelegation(f.sdkCtx, delegation); err != nil {
			panic(err)
		}

		delegation.ValidatorAddress = newValAddr.String()
		// add with new operator addr
		k.SetDelegation(f.sdkCtx, delegation)
	}
}
