package lockup

import (
	"bytes"
	"context"
	"fmt"

	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/defaults/lockup/types"
	stakingtypes "cosmossdk.io/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (bva *BaseLockup) ClawbackFunds(
	ctx context.Context, msg *types.MsgClawback, getLockedCoinsFunc getLockedCoinsFunc,
) (
	*types.MsgClawbackResponse, error,
) {
	// Only allow admin to trigger this operation
	adminAddr, err := bva.addressCodec.StringToBytes(msg.Admin)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid 'admin' address: %s", err)
	}
	admin, err := bva.Admin.Get(ctx)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid admin address: %s", err)
	}
	if !bytes.Equal(adminAddr, admin) {
		return nil, fmt.Errorf("sender is not the admin of this vesting account")
	}

	whoami := accountstd.Whoami(ctx)
	fromAddress, err := bva.addressCodec.BytesToString(whoami)
	if err != nil {
		return nil, err
	}

	clawbackTokens := sdk.Coins{}

	hs := bva.headerService.GetHeaderInfo(ctx)

	lockedCoins, err := getLockedCoinsFunc(ctx, hs.Time, msg.Denoms...)
	if err != nil {
		return nil, err
	}

	// Query bond denom
	paramReq := &stakingtypes.QueryParamsRequest{}
	paramResp, err := accountstd.QueryModule[stakingtypes.QueryParamsResponse](ctx, paramReq)
	if err != nil {
		return nil, err
	}

	for _, denom := range msg.Denoms {
		lockedAmt := lockedCoins.AmountOf(denom)

		// in case of bond denom token, check for scenario when the locked token is being bonded
		// causing the insufficient amount of token to clawback
		if paramResp.Params.BondDenom == denom {
			balance, err := bva.getBalance(ctx, fromAddress, denom)
			if err != nil {
				return nil, err
			}

			spendable, _, err := bva.getSpenableToken(ctx, *balance, sdk.NewCoin(denom, lockedAmt))
			if err != nil {
				return nil, err
			}

			if spendable.AmountOf(denom).LT(lockedAmt) {
				// in case there is not enough token to clawback, proceed to unbond token
				err := bva.forceUnbondLockingDelegations(ctx, fromAddress, paramResp.Params.BondDenom)
				if err != nil {
					return nil, err
				}
				continue
			}

		}
		clawbackTokens = append(clawbackTokens, sdk.NewCoin(denom, lockedAmt))

		// clear the lock token tracking
		err = bva.OriginalLocking.Remove(ctx, denom)
		if err != nil {
			return nil, err
		}
	}
	if len(clawbackTokens) == 0 {
		return nil, fmt.Errorf("no tokens available for clawback")
	}

	// send back to admin
	msgSend := makeMsgSend(fromAddress, string(adminAddr), clawbackTokens)
	_, err = sendMessage(ctx, msgSend)
	if err != nil {
		return nil, err
	}

	return &types.MsgClawbackResponse{}, nil
}

// forceUnbondAllDelegations unbonds all the delegations from the  given account address
func (bva BaseLockup) forceUnbondLockingDelegations(
	ctx context.Context,
	delegator string,
	bondDenom string,
) error {
	// Query account all delegations
	delReq := &stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: delegator,
	}
	delResps, err := accountstd.QueryModule[stakingtypes.QueryDelegatorDelegationsResponse](ctx, delReq)
	if err != nil {
		return err
	}

	for _, resp := range delResps.DelegationResponses {
		del := resp.Delegation

		// unbond
		unbondAmt, err := bva.sk.Unbond(ctx, sdk.AccAddress(del.DelegatorAddress), sdk.ValAddress(del.ValidatorAddress), del.Shares)
		if err != nil {
			return err
		}

		unbondCoins := sdk.Coins{sdk.NewCoin(bondDenom, unbondAmt)}

		err = bva.TrackUndelegation(ctx, unbondCoins)
		if err != nil {
			return err
		}

		isBonded, err := isBonded(ctx, del.DelegatorAddress, del.ValidatorAddress)
		if err != nil {
			return err
		}

		// transfer the validator tokens to the not bonded pool
		if isBonded {
			// doing stakingKeeper.bondedTokensToNotBonded
			err = bva.bk.SendCoinsFromModuleToModule(ctx, stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, unbondCoins)
			if err != nil {
				return err
			}
		}

		err = bva.bk.UndelegateCoinsFromModuleToAccount(ctx, stakingtypes.NotBondedPoolName, sdk.AccAddress(del.DelegatorAddress), unbondCoins)
		if err != nil {
			return err
		}
	}

	return nil
}

func isBonded(ctx context.Context, delAddr, valAddr string) (bool, error) {
	// Query account balance for the sent denom
	req := &stakingtypes.QueryDelegatorValidatorRequest{
		DelegatorAddr: delAddr,
		ValidatorAddr: valAddr,
	}
	resp, err := accountstd.QueryModule[stakingtypes.QueryDelegatorValidatorResponse](ctx, req)
	if err != nil {
		return false, err
	}

	return resp.Validator.IsBonded(), nil
}
