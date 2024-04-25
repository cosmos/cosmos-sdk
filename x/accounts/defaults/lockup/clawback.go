package lockup

import (
	"bytes"
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/defaults/lockup/types"
	banktypes "cosmossdk.io/x/bank/types"

	"time"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	collcodec "cosmossdk.io/collections/codec"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/header"
)

// Compile-time type assertions
var (
	_ types.BaseAccount = (*BaseClawback)(nil)
)

// newBaseClawback creates a new BaseClawback object.
func newBaseClawback(d accountstd.Dependencies) *BaseClawback {
	baseClawback := &BaseClawback{
		Owner:           collections.NewItem(d.SchemaBuilder, types.OwnerPrefix, "owner", collections.BytesValue),
		Admin:           collections.NewItem(d.SchemaBuilder, types.AdminPrefix, "admin", collections.BytesValue),
		OriginalVesting: collections.NewMap(d.SchemaBuilder, types.OriginalVestingPrefix, "original_vesting", collections.StringKey, sdk.IntValue),
		WithdrawedCoins: collections.NewMap(d.SchemaBuilder, types.WithdrawedCoinsPrefix, "withdrawed_coins", collections.StringKey, sdk.IntValue),
		addressCodec:    d.AddressCodec,
		headerService:   d.Environment.HeaderService,
		EndTime:         collections.NewItem(d.SchemaBuilder, types.EndTimePrefix, "end_time", collcodec.KeyToValueCodec[time.Time](sdk.TimeKey)),
	}

	return baseClawback
}

type BaseClawback struct {
	// Owner is the address of the account owner.
	Owner collections.Item[[]byte]
	// Admin is the address who have ability to request lockup account
	// to return the funds
	Admin           collections.Item[[]byte]
	OriginalVesting collections.Map[string, math.Int]
	WithdrawedCoins collections.Map[string, math.Int]
	addressCodec    address.Codec
	headerService   header.Service
	// lockup end time.
	EndTime collections.Item[time.Time]
}

func (bva *BaseClawback) GetEndTime() collections.Item[time.Time] {
	return bva.EndTime
}

func (bva *BaseClawback) GetHeaderService() header.Service {
	return bva.headerService
}

func (bva *BaseClawback) GetOriginalFunds() collections.Map[string, math.Int] {
	return bva.OriginalVesting
}

func (bva *BaseClawback) Init(ctx context.Context, msg *types.MsgInitLockupAccount, amount sdk.Coins) (
	*types.MsgInitLockupAccountResponse, error,
) {
	err := validateMsg(msg, true)
	if err != nil {
		return nil, err
	}

	owner, err := bva.addressCodec.StringToBytes(msg.Owner)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid 'owner' address: %s", err)
	}
	err = bva.Owner.Set(ctx, owner)
	if err != nil {
		return nil, err
	}
	admin, err := bva.addressCodec.StringToBytes(msg.Admin)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid 'admin' address: %s", err)
	}
	err = bva.Admin.Set(ctx, admin)
	if err != nil {
		return nil, err
	}

	funds := accountstd.Funds(ctx)

	// small hack for periodic account init func to pass in funds amount
	if amount != nil && !funds.Equal(amount) {
		return nil, fmt.Errorf("amount need to be equal to funds")
	}

	sortedAmt := funds.Sort()
	for _, coin := range sortedAmt {
		err = bva.OriginalVesting.Set(ctx, coin.Denom, coin.Amount)
		if err != nil {
			return nil, err
		}

		// Set initial value for all locked token
		err = bva.WithdrawedCoins.Set(ctx, coin.Denom, math.ZeroInt())
		if err != nil {
			return nil, err
		}
	}

	err = bva.EndTime.Set(ctx, msg.EndTime)
	if err != nil {
		return nil, err
	}

	return &types.MsgInitLockupAccountResponse{}, nil
}

func (bva *BaseClawback) ClawbackFunds(
	ctx context.Context, msg *types.MsgClawback, getLockedCoinsFunc types.GetLockedCoinsFunc,
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

	for _, denom := range msg.Denoms {
		clawbackAmt := lockedCoins.AmountOf(denom)

		if clawbackAmt.IsZero() {
			continue
		}

		clawbackTokens = append(clawbackTokens, sdk.NewCoin(denom, clawbackAmt))

		// clear the lock token tracking
		err = bva.OriginalVesting.Set(ctx, denom, math.ZeroInt())
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

func (bva *BaseClawback) SendCoins(
	ctx context.Context, msg *types.MsgSend, getLockedCoinsFunc types.GetLockedCoinsFunc,
) (
	*types.MsgExecuteMessagesResponse, error,
) {
	err := bva.checkSender(ctx, msg.Sender)
	if err != nil {
		return nil, err
	}
	whoami := accountstd.Whoami(ctx)
	fromAddress, err := bva.addressCodec.BytesToString(whoami)
	if err != nil {
		return nil, err
	}

	hs := bva.headerService.HeaderInfo(ctx)

	lockedCoins, err := getLockedCoinsFunc(ctx, hs.Time, msg.Amount.Denoms()...)
	if err != nil {
		return nil, err
	}

	err = bva.checkTokensSendable(ctx, fromAddress, msg.Amount, lockedCoins)
	if err != nil {
		return nil, err
	}

	msgSend := &banktypes.MsgSend{
		FromAddress: fromAddress,
		ToAddress:   msg.ToAddress,
		Amount:      msg.Amount,
	}
	responses, err := sendMessage(ctx, msgSend)
	if err != nil {
		return nil, err
	}

	return &types.MsgExecuteMessagesResponse{Responses: responses}, nil
}

// WithdrawUnlockedCoins allow owner to withdraw the unlocked token for a specific denoms to an
// account of choice. Update the withdrawed token tracking for lockup account
func (bva *BaseClawback) WithdrawUnlockedCoins(
	ctx context.Context, msg *types.MsgWithdraw, getLockedCoinsFunc types.GetLockedCoinsFunc,
) (
	*types.MsgWithdrawResponse, error,
) {
	err := bva.checkSender(ctx, msg.Withdrawer)
	if err != nil {
		return nil, err
	}
	whoami := accountstd.Whoami(ctx)
	fromAddress, err := bva.addressCodec.BytesToString(whoami)
	if err != nil {
		return nil, err
	}

	hs := bva.headerService.HeaderInfo(ctx)
	lockedCoins, err := getLockedCoinsFunc(ctx, hs.Time, msg.Denoms...)
	if err != nil {
		return nil, err
	}

	amount := sdk.Coins{}
	for _, denom := range msg.Denoms {
		balance, err := getBalance(ctx, fromAddress, denom)
		if err != nil {
			return nil, err
		}
		lockedAmt := lockedCoins.AmountOf(denom)

		spendable, err := balance.SafeSub(sdk.NewCoin(denom, lockedAmt))
		if err != nil {
			return nil, errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds,
				"locked amount exceeds account balance funds")
		}

		withdrawedAmt, err := bva.WithdrawedCoins.Get(ctx, denom)
		if err != nil {
			return nil, err
		}
		originalLockingAmt, err := bva.OriginalVesting.Get(ctx, denom)
		if err != nil {
			return nil, err
		}

		// withdrawable amount is equal to original locking amount subtract already withdrawed amount
		withdrawableAmt, err := originalLockingAmt.SafeSub(withdrawedAmt)
		if err != nil {
			return nil, err
		}

		withdrawAmt := math.MinInt(withdrawableAmt, spendable.Amount)
		// if zero amount go to the next iteration
		if withdrawAmt.IsZero() {
			continue
		}
		amount = append(amount, sdk.NewCoin(denom, withdrawAmt))

		// update the withdrawed amount
		err = bva.WithdrawedCoins.Set(ctx, denom, withdrawedAmt.Add(withdrawAmt))
		if err != nil {
			return nil, err
		}
	}
	if len(amount) == 0 {
		return nil, fmt.Errorf("no tokens available for withdrawing")
	}

	msgSend := &banktypes.MsgSend{
		FromAddress: fromAddress,
		ToAddress:   msg.ToAddress,
		Amount:      amount,
	}
	_, err = sendMessage(ctx, msgSend)
	if err != nil {
		return nil, err
	}

	return &types.MsgWithdrawResponse{
		Receiver:       msg.ToAddress,
		AmountReceived: amount,
	}, nil
}

func (bva *BaseClawback) checkSender(ctx context.Context, sender string) error {
	owner, err := bva.Owner.Get(ctx)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid owner address: %s", err.Error())
	}
	senderBytes, err := bva.addressCodec.StringToBytes(sender)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err.Error())
	}
	if !bytes.Equal(owner, senderBytes) {
		return fmt.Errorf("sender is not the owner of this vesting account")
	}

	return nil
}

func (bva BaseClawback) checkTokensSendable(ctx context.Context, sender string, amount, lockedCoins sdk.Coins) error {
	// Check if any sent tokens is exceeds lockup account balances
	for _, coin := range amount {
		balance, err := getBalance(ctx, sender, coin.Denom)
		if err != nil {
			return err
		}
		lockedAmt := lockedCoins.AmountOf(coin.Denom)

		spendable, hasNeg := sdk.Coins{*balance}.SafeSub(sdk.NewCoin(coin.Denom, lockedAmt))
		if hasNeg {
			return errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds,
				"locked amount exceeds account balance funds")
		}

		if _, hasNeg := spendable.SafeSub(coin); hasNeg {
			return errorsmod.Wrapf(
				sdkerrors.ErrInsufficientFunds,
				"spendable balance %s is smaller than %s",
				spendable, coin,
			)
		}
	}

	return nil
}

// QueryAccountBaseInfo returns a lockup account's info
func (bva BaseClawback) QueryAccountBaseInfo(ctx context.Context, _ *types.QueryLockupAccountInfoRequest) (
	*types.QueryLockupAccountInfoResponse, error,
) {
	owner, err := bva.Owner.Get(ctx)
	if err != nil {
		return nil, err
	}

	ownerAddress, err := bva.addressCodec.BytesToString(owner)
	if err != nil {
		return nil, err
	}

	endTime, err := bva.EndTime.Get(ctx)
	if err != nil {
		return nil, err
	}

	originalLocking := sdk.Coins{}
	err = IterateCoinEntries(ctx, bva.OriginalVesting, func(key string, value math.Int) (stop bool, err error) {
		originalLocking = append(originalLocking, sdk.NewCoin(key, value))
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return &types.QueryLockupAccountInfoResponse{
		Owner:           ownerAddress,
		OriginalLocking: originalLocking,
		EndTime:         &endTime,
	}, nil
}
