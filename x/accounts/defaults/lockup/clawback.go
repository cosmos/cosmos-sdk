package lockup

import (
	"bytes"
	"context"
	"fmt"

	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/defaults/lockup/types"
	banktypes "cosmossdk.io/x/bank/types"
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

	hs := bva.headerService.HeaderInfo(ctx)

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
		var clawbackAmt math.Int
		lockedAmt := lockedCoins.AmountOf(denom)
		clawbackAmt = lockedAmt

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

			spendableAmt := spendable.AmountOf(denom)

			if spendableAmt.LT(lockedAmt) {
				// in case there is not enough token to clawback, proceed to unbond token
				err := bva.forceUnbondLockingDelegations(ctx, fromAddress, paramResp.Params.BondDenom)
				if err != nil {
					return nil, err
				}

				deptAmt, err := lockedAmt.SafeSub(spendableAmt)
				if err != nil {
					return nil, err
				}

				// clawback the available amount first
				clawbackAmt = spendableAmt

				// track the remain amount
				bva.ClawbackDept.Set(ctx, denom, deptAmt)
			}

		}
		clawbackTokens = append(clawbackTokens, sdk.NewCoin(denom, clawbackAmt))

		// clear the lock token tracking
		err = bva.OriginalLocking.Set(ctx, denom, math.ZeroInt())
		if err != nil {
			return nil, err
		}
	}
	if len(clawbackTokens) == 0 {
		return nil, fmt.Errorf("no tokens available for clawback")
	}

	// send back to admin
	msgSend := &banktypes.MsgSend{
		FromAddress: fromAddress,
		ToAddress:   msg.Admin,
		Amount:      clawbackTokens,
	}
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

		val, err := getVal(ctx, del.DelegatorAddress, del.ValidatorAddress)
		if err != nil {
			return err
		}

		delAmt := val.TokensFromShares(del.Shares)
		delCoin := sdk.NewCoin(bondDenom, delAmt.TruncateInt())

		err = bva.TrackUndelegation(ctx, sdk.Coins{delCoin})
		if err != nil {
			return err
		}

		msgUndelegate := &stakingtypes.MsgUndelegate{
			DelegatorAddress: delegator,
			ValidatorAddress: del.ValidatorAddress,
			Amount:           delCoin,
		}
		_, err = sendMessage(ctx, msgUndelegate)
		if err != nil {
			return err
		}
	}

	return nil
}

func getVal(ctx context.Context, delAddr, valAddr string) (stakingtypes.Validator, error) {
	// Query account balance for the sent denom
	req := &stakingtypes.QueryDelegatorValidatorRequest{
		DelegatorAddr: delAddr,
		ValidatorAddr: valAddr,
	}
	resp, err := accountstd.QueryModule[stakingtypes.QueryDelegatorValidatorResponse](ctx, req)
	if err != nil {
		return stakingtypes.Validator{}, err
	}

	return resp.Validator, nil
}
