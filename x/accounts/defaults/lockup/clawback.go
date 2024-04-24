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

	"bytes"
	"context"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	collcodec "cosmossdk.io/collections/codec"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/header"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/defaults/lockup/types"

	banktypes "cosmossdk.io/x/bank/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type getLockedCoinsFunc = func(ctx context.Context, time time.Time, denoms ...string) (sdk.Coins, error)

// newBaseLockup creates a new BaseLockup object.
func newBaseLockup(d accountstd.Dependencies) *BaseLockup {
	BaseLockup := &BaseLockup{
		Owner:            collections.NewItem(d.SchemaBuilder, OwnerPrefix, "owner", collections.BytesValue),
		Admin:            collections.NewItem(d.SchemaBuilder, AdminPrefix, "admin", collections.BytesValue),
		ClawbackDebt:     collections.NewMap(d.SchemaBuilder, ClawbackDebtPrefix, "clawback_debt", collections.StringKey, sdk.IntValue),
		OriginalLocking:  collections.NewMap(d.SchemaBuilder, OriginalLockingPrefix, "original_locking", collections.StringKey, sdk.IntValue),
		DelegatedFree:    collections.NewMap(d.SchemaBuilder, DelegatedFreePrefix, "delegated_free", collections.StringKey, sdk.IntValue),
		DelegatedLocking: collections.NewMap(d.SchemaBuilder, DelegatedLockingPrefix, "delegated_locking", collections.StringKey, sdk.IntValue),
		WithdrawedCoins:  collections.NewMap(d.SchemaBuilder, WithdrawedCoinsPrefix, "withdrawed_coins", collections.StringKey, sdk.IntValue),
		addressCodec:     d.AddressCodec,
		headerService:    d.Environment.HeaderService,
		EndTime:          collections.NewItem(d.SchemaBuilder, EndTimePrefix, "end_time", collcodec.KeyToValueCodec[time.Time](sdk.TimeKey)),
	}

	return BaseLockup
}

type BaseLockup struct {
	// Owner is the address of the account owner.
	Owner collections.Item[[]byte]
	// Admin is the address who have ability to request lockup account
	// to return the funds
	Admin            collections.Item[[]byte]
	ClawbackDebt     collections.Map[string, math.Int]
	OriginalLocking  collections.Map[string, math.Int]
	DelegatedFree    collections.Map[string, math.Int]
	DelegatedLocking collections.Map[string, math.Int]
	WithdrawedCoins  collections.Map[string, math.Int]
	addressCodec     address.Codec
	headerService    header.Service
	// lockup end time.
	EndTime collections.Item[time.Time]
}

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

			balanceAmt := balance.Amount

			clawbackDebtAmt, err := bva.OriginalLocking.Get(ctx, denom)
			if err != nil {
				return nil, err
			}

			// if clawback debt exist which mean the clawback process for this denom had been triggered
			// handle the debt if the balance is sufficient
			if !clawbackDebtAmt.IsZero() && balanceAmt.GT(clawbackDebtAmt) {
				clawbackTokens = append(clawbackTokens, sdk.NewCoin(denom, clawbackDebtAmt))

				// clear the debt tracking
				err = bva.ClawbackDebt.Set(ctx, denom, math.ZeroInt())
				if err != nil {
					return nil, err
				}
				continue
			}

			// check if balance is sufficient
			if balanceAmt.LT(lockedAmt) {
				// in case there is not enough token to clawback, proceed to unbond token
				err := bva.forceUnbondDelegations(ctx, fromAddress, paramResp.Params.BondDenom)
				if err != nil {
					return nil, err
				}

				debtAmt, err := lockedAmt.SafeSub(balanceAmt)
				if err != nil {
					return nil, err
				}

				// clawback the available amount first
				clawbackAmt = balanceAmt

				// track the remain amount
				err = bva.ClawbackDebt.Set(ctx, denom, debtAmt)
				if err != nil {
					return nil, err
				}
			}

		}

		if clawbackAmt.IsZero() {
			continue
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
func (bva BaseLockup) forceUnbondDelegations(
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
