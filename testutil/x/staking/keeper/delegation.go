package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetAllDelegations returns all delegations used during genesis dump.
func (k Keeper) GetAllDelegations(ctx context.Context) ([]types.Delegation, error) {
	var delegations types.Delegations
	err := k.Delegations.Walk(ctx, nil,
		func(key collections.Pair[sdk.AccAddress, sdk.ValAddress], delegation types.Delegation) (stop bool, err error) {
			delegations = append(delegations, delegation)
			return false, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return delegations, nil
}

// SetDelegation sets a delegation.
func (k Keeper) SetDelegation(ctx context.Context, delegation types.Delegation) error {
	delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(delegation.DelegatorAddress)
	if err != nil {
		return err
	}

	valAddr, err := k.validatorAddressCodec.StringToBytes(delegation.GetValidatorAddr())
	if err != nil {
		return err
	}

	err = k.Delegations.Set(ctx, collections.Join(sdk.AccAddress(delegatorAddress), sdk.ValAddress(valAddr)), delegation)
	if err != nil {
		return err
	}

	return nil
}

// Delegate performs a delegation, set/update everything necessary within the store.
// tokenSrc indicates the bond status of the incoming funds.
func (k Keeper) Delegate(
	ctx context.Context, delAddr sdk.AccAddress, bondAmt math.Int, tokenSrc types.BondStatus,
	validator types.Validator, subtractAccount bool,
) (newShares math.LegacyDec, err error) {
	// In some situations, the exchange rate becomes invalid, e.g. if
	// Validator loses all tokens due to slashing. In this case,
	// make all future delegations invalid.
	if validator.InvalidExRate() {
		return math.LegacyZeroDec(), types.ErrDelegatorShareExRateInvalid
	}

	valbz, err := k.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
	if err != nil {
		return math.LegacyZeroDec(), err
	}

	// Get or create the delegation object and call the appropriate hook if present
	delegation, err := k.Delegations.Get(ctx, collections.Join(delAddr, sdk.ValAddress(valbz)))
	if err == nil {
		// found
		err = k.Hooks().BeforeDelegationSharesModified(ctx, delAddr, valbz)
	} else if errors.Is(err, collections.ErrNotFound) {
		// not found
		delAddrStr, err1 := k.authKeeper.AddressCodec().BytesToString(delAddr)
		if err1 != nil {
			return math.LegacyDec{}, err1
		}

		delegation = types.NewDelegation(delAddrStr, validator.GetOperator(), math.LegacyZeroDec())
		err = k.Hooks().BeforeDelegationCreated(ctx, delAddr, valbz)
	} else {
		return math.LegacyZeroDec(), err
	}

	if err != nil {
		return math.LegacyZeroDec(), err
	}

	// if subtractAccount is true then we are
	// performing a delegation and not a redelegation, thus the source tokens are
	// all non bonded
	if subtractAccount {
		if tokenSrc == types.Bonded {
			return math.LegacyZeroDec(), fmt.Errorf("delegation token source cannot be bonded; expected Unbonded or Unbonding, got Bonded")
		}

		var sendName string

		switch {
		case validator.IsBonded():
			sendName = types.BondedPoolName
		case validator.IsUnbonding(), validator.IsUnbonded():
			sendName = types.NotBondedPoolName
		default:
			return math.LegacyZeroDec(), fmt.Errorf("invalid validator status: %v", validator.Status)
		}

		bondDenom, err := k.BondDenom(ctx)
		if err != nil {
			return math.LegacyDec{}, err
		}

		coins := sdk.NewCoins(sdk.NewCoin(bondDenom, bondAmt))
		if err := k.bankKeeper.DelegateCoinsFromAccountToModule(ctx, delAddr, sendName, coins); err != nil {
			return math.LegacyDec{}, err
		}
	} else {
		// potentially transfer tokens between pools, if
		switch {
		case tokenSrc == types.Bonded && validator.IsBonded():
			// do nothing
		case (tokenSrc == types.Unbonded || tokenSrc == types.Unbonding) && !validator.IsBonded():
			// do nothing
		case (tokenSrc == types.Unbonded || tokenSrc == types.Unbonding) && validator.IsBonded():
			// transfer pools
			err = k.notBondedTokensToBonded(ctx, bondAmt)
			if err != nil {
				return math.LegacyDec{}, err
			}
		case tokenSrc == types.Bonded && !validator.IsBonded():
			// transfer pools
			err = k.bondedTokensToNotBonded(ctx, bondAmt)
			if err != nil {
				return math.LegacyDec{}, err
			}
		default:
			return math.LegacyZeroDec(), fmt.Errorf("unknown token source bond status: %v", tokenSrc)
		}
	}

	_, newShares, err = k.AddValidatorTokensAndShares(ctx, validator, bondAmt)
	if err != nil {
		return newShares, err
	}

	// Update delegation
	delegation.Shares = delegation.Shares.Add(newShares)
	if err = k.SetDelegation(ctx, delegation); err != nil {
		return newShares, err
	}

	// Call the after-modification hook
	if err := k.Hooks().AfterDelegationModified(ctx, delAddr, valbz); err != nil {
		return newShares, err
	}

	return newShares, nil
}

// Unbond unbonds a particular delegation and perform associated store operations.
func (k Keeper) Unbond(
	ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, shares math.LegacyDec,
) (amount math.Int, err error) {
	// TODO: implement this if needed
	return
}
