package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// RegisterInvariants registers all governance invariants
func RegisterInvariants(ir sdk.InvariantRegistry, keeper *Keeper, bk types.BankKeeper) {
	ir.RegisterRoute(types.ModuleName, "module-account", ModuleAccountInvariant(keeper, bk))
	ir.RegisterRoute(types.ModuleName, "governors-delegations", GovernorsDelegationsInvariant(keeper, keeper.sk))
}

// ModuleAccountInvariant checks that the module account coins reflects the sum of
// deposit amounts held on store.
func ModuleAccountInvariant(keeper *Keeper, bk types.BankKeeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var expectedDeposits sdk.Coins

		err := keeper.Deposits.Walk(ctx, nil, func(_ collections.Pair[uint64, sdk.AccAddress], value v1.Deposit) (stop bool, err error) {
			expectedDeposits = expectedDeposits.Add(value.Amount...)
			return false, nil
		})
		if err != nil {
			panic(err)
		}

		macc := keeper.GetGovernanceAccount(ctx)
		balances := bk.GetAllBalances(ctx, macc.GetAddress())

		// Require that the deposit balances are <= than the x/gov module's total
		// balances. We use the <= operator since external funds can be sent to x/gov
		// module's account and so the balance can be larger.
		broken := !balances.IsAllGTE(expectedDeposits)

		return sdk.FormatInvariant(types.ModuleName, "deposits",
			fmt.Sprintf("\tgov ModuleAccount coins: %s\n\tsum of deposit amounts:  %s\n",
				balances, expectedDeposits)), broken
	}
}

// GovernorsDelegationsInvariant checks that the validator shares resulting from all
// governor delegations actually correspond to the stored governor validator shares
func GovernorsDelegationsInvariant(keeper *Keeper, sk types.StakingKeeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken       = false
			invariantStr string
		)

		// keeper.IterateGovernors(ctx, func(index int64, governor v1.GovernorI) bool {
		keeper.Governors.Walk(ctx, nil, func(_ types.GovernorAddress, governor v1.Governor) (stop bool, err error) {
			// check that if governor is active, it has a valid governance self-delegation
			if governor.IsActive() {
				if del, err := keeper.GovernanceDelegations.Get(ctx, sdk.AccAddress(governor.GetAddress())); err != nil || !governor.GetAddress().Equals(types.MustGovernorAddressFromBech32(del.GovernorAddress)) {
					invariantStr = sdk.FormatInvariant(types.ModuleName, fmt.Sprintf("governor %s delegations", governor.GetAddress().String()),
						"active governor without governance self-delegation")
					broken = true
					return true, nil
				}
			}

			valShares := make(map[string]math.LegacyDec)
			valSharesKeys := make([]string, 0)
			err = keeper.GovernanceDelegationsByGovernor.Walk(ctx, collections.NewPrefixedPairRange[types.GovernorAddress, sdk.AccAddress](governor.GetAddress()), func(_ collections.Pair[types.GovernorAddress, sdk.AccAddress], delegation v1.GovernanceDelegation) (stop bool, err error) {
				delAddr := sdk.MustAccAddressFromBech32(delegation.DelegatorAddress)
				err = sk.IterateDelegations(ctx, delAddr, func(_ int64, delegation stakingtypes.DelegationI) (stop bool) {
					validatorAddr := delegation.GetValidatorAddr()
					shares := delegation.GetShares()
					if _, ok := valShares[validatorAddr]; !ok {
						valShares[validatorAddr] = math.LegacyZeroDec()
						valSharesKeys = append(valSharesKeys, validatorAddr)
					}
					valShares[validatorAddr] = valShares[validatorAddr].Add(shares)
					return false
				})
				if err != nil {
					invariantStr = sdk.FormatInvariant(types.ModuleName, fmt.Sprintf("governor %s delegations", governor.GetAddress().String()),
						fmt.Sprintf("failed to iterate delegations: %v", err))
					broken = true
					return true, nil
				}
				return false, nil
			})
			if err != nil {
				invariantStr = sdk.FormatInvariant(types.ModuleName, fmt.Sprintf("governor %s delegations", governor.GetAddress().String()),
					fmt.Sprintf("failed to iterate governance delegations: %v", err))
				broken = true
				return true, nil
			}

			for _, valAddrStr := range valSharesKeys {
				shares := valShares[valAddrStr]
				validatorAddr, _ := sdk.ValAddressFromBech32(valAddrStr)
				vs, err := keeper.ValidatorSharesByGovernor.Get(ctx, collections.Join(governor.GetAddress(), validatorAddr))
				if err != nil {
					invariantStr = sdk.FormatInvariant(types.ModuleName, fmt.Sprintf("governor %s delegations", governor.GetAddress().String()),
						fmt.Sprintf("validator %s shares not found or unable to retrieve them", valAddrStr))
					broken = true
					return true, nil
				}
				if !vs.Shares.Equal(shares) {
					invariantStr = sdk.FormatInvariant(types.ModuleName, fmt.Sprintf("governor %s delegations", governor.GetAddress().String()),
						fmt.Sprintf("stored shares %s for validator %s do not match actual shares %s", vs.Shares, valAddrStr, shares))
					broken = true
					return true, nil
				}
			}

			keeper.ValidatorSharesByGovernor.Walk(ctx, nil, func(_ collections.Pair[types.GovernorAddress, sdk.ValAddress], shares v1.GovernorValShares) (stop bool, err error) {
				if _, ok := valShares[shares.ValidatorAddress]; !ok && shares.Shares.GT(math.LegacyZeroDec()) {
					invariantStr = sdk.FormatInvariant(types.ModuleName, fmt.Sprintf("governor %s delegations", governor.GetAddress().String()),
						fmt.Sprintf("non-zero (%s) shares stored for validator %s where there should be none", shares.Shares, shares.ValidatorAddress))
					broken = true
					return true, nil
				}
				return false, nil
			})

			return broken, nil
		})
		return invariantStr, broken
	}
}
