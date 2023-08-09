package keeper_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"cosmossdk.io/collections"
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
		if err := f.stakingKeeper.SetValidator(f.sdkCtx, validator); err != nil {
			panic(err)
		}
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
		if err := f.stakingKeeper.SetValidator(f.sdkCtx, validator); err != nil {
			panic(err)
		}
	}

	delegationsNum := 1000
	for _, val := range valAddrs {
		for i := 0; i < delegationsNum; i++ {
			delegator := sdk.AccAddress(fmt.Sprintf("address%d", i))
			err := banktestutil.FundAccount(f.sdkCtx, f.bankKeeper, delegator,
				sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(int64(i)))))
			if err != nil {
				panic(err)
			}
			NewDel := types.NewDelegation(delegator.String(), val.String(), math.LegacyNewDec(int64(i)))

			if err := f.stakingKeeper.SetDelegation(f.sdkCtx, NewDel); err != nil {
				panic(err)
			}
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
		if err := f.stakingKeeper.SetValidator(f.sdkCtx, validator); err != nil {
			panic(err)
		}
	}

	delegationsNum := 1000
	for _, val := range valAddrs {
		for i := 0; i < delegationsNum; i++ {
			delegator := sdk.AccAddress(fmt.Sprintf("address%d", i))
			err := banktestutil.FundAccount(f.sdkCtx, f.bankKeeper, delegator,
				sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(int64(i)))))
			if err != nil {
				panic(err)
			}
			NewDel := types.NewDelegation(delegator.String(), val.String(), math.LegacyNewDec(int64(i)))
			if err := f.stakingKeeper.SetDelegation(f.sdkCtx, NewDel); err != nil {
				panic(err)
			}
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
		valAddr, err := k.ValidatorAddressCodec().StringToBytes(delegation.GetValidatorAddr())
		if err != nil {
			panic(err)
		}

		if bytes.EqualFold(valAddr, existingValAddr) {
			if err := k.RemoveDelegation(f.sdkCtx, delegation); err != nil {
				panic(err)
			}
			delegation.ValidatorAddress = newValAddr.String()
			if err := k.SetDelegation(f.sdkCtx, delegation); err != nil {
				panic(err)
			}
		}
	}
}

func updateValidatorDelegations(f *fixture, existingValAddr, newValAddr sdk.ValAddress) {
	k := f.stakingKeeper

	rng := collections.NewPrefixedPairRange[sdk.ValAddress, sdk.AccAddress](existingValAddr)
	err := k.DelegationsByValidator.Walk(f.sdkCtx, rng, func(key collections.Pair[sdk.ValAddress, sdk.AccAddress], _ []byte) (stop bool, err error) {
		valAddr, delAddr := key.K1(), key.K2()

		delegation, err := k.Delegations.Get(f.sdkCtx, collections.Join(delAddr, valAddr))
		if err != nil {
			return true, err
		}

		// remove old operator addr from delegation
		if err := k.RemoveDelegation(f.sdkCtx, delegation); err != nil {
			return true, err
		}

		delegation.ValidatorAddress = newValAddr.String()
		// add with new operator addr
		if err := k.SetDelegation(f.sdkCtx, delegation); err != nil {
			return true, err
		}

		return false, nil
	})
	if err != nil && !errors.Is(err, collections.ErrInvalidIterator) {
		panic(err)
	}
}
