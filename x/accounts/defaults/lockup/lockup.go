package lockup

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/header"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	lockuptypes "cosmossdk.io/x/accounts/lockup/types"
	banktypes "cosmossdk.io/x/bank/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/gogoproto/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	OriginalLockingPrefix  = collections.NewPrefix(0)
	DelegatedFreePrefix    = collections.NewPrefix(1)
	DelegatedLockingPrefix = collections.NewPrefix(2)
	EndTimePrefix          = collections.NewPrefix(3)
	StartTimePrefix        = collections.NewPrefix(4)
	LockingPeriodsPrefix   = collections.NewPrefix(5)
	OwnerPrefix            = collections.NewPrefix(6)
	WithdrawedCoinsPrefix  = collections.NewPrefix(7)
)

var (
	CONTINUOUS_LOCKING_ACCOUNT = "continuous-locking-account"
	DELAYED_LOCKING_ACCOUNT    = "delayed-locking-account"
	PERIODIC_LOCKING_ACCOUNT   = "periodic-locking-account"
	PERMANENT_LOCKING_ACCOUNT  = "permanent-locking-account"
)

type getLockedCoinsFunc = func(ctx context.Context, time time.Time, denoms ...string) (sdk.Coins, error)

// newBaseLockup creates a new BaseLockup object.
func newBaseLockup(d accountstd.Dependencies) *BaseLockup {
	BaseLockup := &BaseLockup{
		Owner:            collections.NewItem(d.SchemaBuilder, OwnerPrefix, "owner", collections.BytesValue),
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
	Owner            collections.Item[[]byte]
	OriginalLocking  collections.Map[string, math.Int]
	DelegatedFree    collections.Map[string, math.Int]
	DelegatedLocking collections.Map[string, math.Int]
	WithdrawedCoins  collections.Map[string, math.Int]
	addressCodec     address.Codec
	headerService    header.Service
	// lockup end time.
	EndTime collections.Item[time.Time]
}

func (bva *BaseLockup) Init(ctx context.Context, msg *lockuptypes.MsgInitLockupAccount) (
	*lockuptypes.MsgInitLockupAccountResponse, error,
) {
	owner, err := bva.addressCodec.StringToBytes(msg.Owner)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid 'owner' address: %s", err)
	}
	err = bva.Owner.Set(ctx, owner)
	if err != nil {
		return nil, err
	}

	funds := accountstd.Funds(ctx)

	sortedAmt := funds.Sort()
	for _, coin := range sortedAmt {
		err = bva.OriginalLocking.Set(ctx, coin.Denom, coin.Amount)
		if err != nil {
			return nil, err
		}

		// Set initial value for all locked token
		err = bva.WithdrawedCoins.Set(ctx, coin.Denom, math.ZeroInt())
		if err != nil {
			return nil, err
		}

		// Set initial value for all locked token
		err = bva.DelegatedFree.Set(ctx, coin.Denom, math.ZeroInt())
		if err != nil {
			return nil, err
		}

		// Set initial value for all locked token
		err = bva.DelegatedLocking.Set(ctx, coin.Denom, math.ZeroInt())
		if err != nil {
			return nil, err
		}
	}

	err = bva.EndTime.Set(ctx, msg.EndTime)
	if err != nil {
		return nil, err
	}

	return &lockuptypes.MsgInitLockupAccountResponse{}, nil
}

func (bva *BaseLockup) Delegate(
	ctx context.Context, msg *lockuptypes.MsgDelegate, getLockedCoinsFunc getLockedCoinsFunc,
) (
	*lockuptypes.MsgExecuteMessagesResponse, error,
) {
	err := bva.checkSender(ctx, msg.Sender)
	if err != nil {
		return nil, err
	}
	whoami := accountstd.Whoami(ctx)
	delegatorAddress, err := bva.addressCodec.BytesToString(whoami)
	if err != nil {
		return nil, err
	}

	hs := bva.headerService.GetHeaderInfo(ctx)

	balance, err := bva.getBalance(ctx, delegatorAddress, msg.Amount.Denom)
	if err != nil {
		return nil, err
	}
	lockedCoins, err := getLockedCoinsFunc(ctx, hs.Time, msg.Amount.Denom)
	if err != nil {
		return nil, err
	}

	err = bva.TrackDelegation(
		ctx,
		sdk.Coins{*balance},
		lockedCoins,
		sdk.Coins{msg.Amount},
	)
	if err != nil {
		return nil, err
	}

	msgDelegate := makeMsgDelegate(delegatorAddress, msg.ValidatorAddress, msg.Amount)
	responses, err := sendMessage(ctx, msgDelegate)
	if err != nil {
		return nil, err
	}

	return &lockuptypes.MsgExecuteMessagesResponse{Responses: responses}, nil
}

func (bva *BaseLockup) Undelegate(
	ctx context.Context, msg *lockuptypes.MsgUndelegate,
) (
	*lockuptypes.MsgExecuteMessagesResponse, error,
) {
	err := bva.checkSender(ctx, msg.Sender)
	if err != nil {
		return nil, err
	}
	whoami := accountstd.Whoami(ctx)
	delegatorAddress, err := bva.addressCodec.BytesToString(whoami)
	if err != nil {
		return nil, err
	}

	err = bva.TrackUndelegation(ctx, sdk.Coins{msg.Amount})
	if err != nil {
		return nil, err
	}

	msgUndelegate := makeMsgUndelegate(delegatorAddress, msg.ValidatorAddress, msg.Amount)
	responses, err := sendMessage(ctx, msgUndelegate)
	if err != nil {
		return nil, err
	}

	return &lockuptypes.MsgExecuteMessagesResponse{Responses: responses}, nil
}

func (bva *BaseLockup) SendCoins(
	ctx context.Context, msg *lockuptypes.MsgSend, getLockedCoinsFunc getLockedCoinsFunc,
) (
	*lockuptypes.MsgExecuteMessagesResponse, error,
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

	hs := bva.headerService.GetHeaderInfo(ctx)

	lockedCoins, err := getLockedCoinsFunc(ctx, hs.Time, msg.Amount.Denoms()...)
	if err != nil {
		return nil, err
	}

	err = bva.checkTokensSendable(ctx, fromAddress, msg.Amount, lockedCoins)
	if err != nil {
		return nil, err
	}

	msgSend := makeMsgSend(fromAddress, msg.ToAddress, msg.Amount)
	responses, err := sendMessage(ctx, msgSend)
	if err != nil {
		return nil, err
	}

	return &lockuptypes.MsgExecuteMessagesResponse{Responses: responses}, nil
}

// WithdrawUnlockedCoins allow owner to withdraw the unlocked token for a specific denoms to an
// account of choice. Update the withdrawed token tracking for lockup account
func (bva *BaseLockup) WithdrawUnlockedCoins(
	ctx context.Context, msg *lockuptypes.MsgWithdraw, getLockedCoinsFunc getLockedCoinsFunc,
) (
	*lockuptypes.MsgWithdrawResponse, error,
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

	hs := bva.headerService.GetHeaderInfo(ctx)
	lockedCoins, err := getLockedCoinsFunc(ctx, hs.Time, msg.Denoms...)
	if err != nil {
		return nil, err
	}

	amount := sdk.Coins{}
	for _, denom := range msg.Denoms {
		balance, err := bva.getBalance(ctx, fromAddress, denom)
		if err != nil {
			return nil, err
		}
		lockedAmt := lockedCoins.AmountOf(denom)

		// get lockedCoin from that are not bonded for the sent denom
		notBondedLockedCoin, err := bva.GetNotBondedLockedCoin(ctx, sdk.NewCoin(denom, lockedAmt), denom)
		if err != nil {
			return nil, err
		}

		spendable, err := balance.SafeSub(notBondedLockedCoin)
		if err != nil {
			return nil, errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds,
				"locked amount exceeds account balance funds: %s > %s", notBondedLockedCoin, balance)
		}

		withdrawedAmt, err := bva.WithdrawedCoins.Get(ctx, denom)
		if err != nil {
			return nil, err
		}
		originalLockingAmt, err := bva.OriginalLocking.Get(ctx, denom)
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

	msgSend := makeMsgSend(fromAddress, msg.ToAddress, amount)
	_, err = sendMessage(ctx, msgSend)
	if err != nil {
		return nil, err
	}

	return &lockuptypes.MsgWithdrawResponse{
		Receiver:       msg.ToAddress,
		AmountReceived: amount,
	}, nil
}

func (bva *BaseLockup) checkSender(ctx context.Context, sender string) error {
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

func sendMessage(ctx context.Context, msg proto.Message) ([]*codectypes.Any, error) {
	response, err := accountstd.ExecModuleUntyped(ctx, msg)
	if err != nil {
		return nil, err
	}

	respAny, err := accountstd.PackAny(response)
	if err != nil {
		return nil, err
	}

	return []*codectypes.Any{respAny}, nil
}

// TrackDelegation tracks a delegation amount for any given lockup account type
// given the amount of coins currently being locked and the current account balance
// of the delegation denominations.
//
// CONTRACT: The account's coins, delegation coins, locked coins, and delegated
// locking coins must be sorted.
func (bva *BaseLockup) TrackDelegation(
	ctx context.Context, balance, lockedCoins, amount sdk.Coins,
) error {
	for _, coin := range amount {
		baseAmt := balance.AmountOf(coin.Denom)
		lockedAmt := lockedCoins.AmountOf(coin.Denom)
		delLockingAmt, err := bva.DelegatedLocking.Get(ctx, coin.Denom)
		if err != nil {
			return err
		}
		delFreeAmt, err := bva.DelegatedFree.Get(ctx, coin.Denom)
		if err != nil {
			return err
		}

		// return error if the delegation amount is zero or if the base coins does not
		// exceed the desired delegation amount.
		if coin.Amount.IsZero() || baseAmt.LT(coin.Amount) {
			return sdkerrors.ErrInvalidCoins.Wrap("delegation attempt with zero coins or insufficient funds")
		}

		// compute x and y per the specification, where:
		// X := min(max(V - DV, 0), D)
		// Y := D - X
		x := math.MinInt(math.MaxInt(lockedAmt.Sub(delLockingAmt), math.ZeroInt()), coin.Amount)
		y := coin.Amount.Sub(x)

		delLockingCoin := sdk.NewCoin(coin.Denom, delLockingAmt)
		delFreeCoin := sdk.NewCoin(coin.Denom, delFreeAmt)
		if !x.IsZero() {
			xCoin := sdk.NewCoin(coin.Denom, x)
			newDelLocking := delLockingCoin.Add(xCoin)
			err = bva.DelegatedLocking.Set(ctx, newDelLocking.Denom, newDelLocking.Amount)
			if err != nil {
				return err
			}
		}

		if !y.IsZero() {
			yCoin := sdk.NewCoin(coin.Denom, y)
			newDelFree := delFreeCoin.Add(yCoin)
			err = bva.DelegatedFree.Set(ctx, newDelFree.Denom, newDelFree.Amount)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// TrackUndelegation tracks an undelegation amount by setting the necessary
// values by which delegated locking and delegated free need to decrease and
// by which amount the base coins need to increase.
//
// NOTE: The undelegation (bond refund) amount may exceed the delegated
// locking (bond) amount due to the way undelegation truncates the bond refund,
// which can increase the validator's exchange rate (tokens/shares) slightly if
// the undelegated tokens are non-integral.
//
// CONTRACT: The account's coins and undelegation coins must be sorted.
func (bva *BaseLockup) TrackUndelegation(ctx context.Context, amount sdk.Coins) error {
	for _, coin := range amount {
		// return error if the undelegation amount is zero
		if coin.Amount.IsZero() {
			return sdkerrors.ErrInvalidCoins.Wrap("undelegation attempt with zero coins")
		}
		delFreeAmt, err := bva.DelegatedFree.Get(ctx, coin.Denom)
		if err != nil {
			return err
		}
		delLockingAmt, err := bva.DelegatedLocking.Get(ctx, coin.Denom)
		if err != nil {
			return err
		}

		// compute x and y per the specification, where:
		// X := min(DF, D)
		// Y := min(DV, D - X)
		x := math.MinInt(delFreeAmt, coin.Amount)
		y := math.MinInt(delLockingAmt, coin.Amount.Sub(x))

		delLockingCoin := sdk.NewCoin(coin.Denom, delLockingAmt)
		delFreeCoin := sdk.NewCoin(coin.Denom, delFreeAmt)
		if !x.IsZero() {
			xCoin := sdk.NewCoin(coin.Denom, x)
			newDelFree := delFreeCoin.Sub(xCoin)
			err = bva.DelegatedFree.Set(ctx, newDelFree.Denom, newDelFree.Amount)
			if err != nil {
				return err
			}
		}

		if !y.IsZero() {
			yCoin := sdk.NewCoin(coin.Denom, y)
			newDelLocking := delLockingCoin.Sub(yCoin)
			err = bva.DelegatedLocking.Set(ctx, newDelLocking.Denom, newDelLocking.Amount)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (bva BaseLockup) getBalance(ctx context.Context, sender, denom string) (*sdk.Coin, error) {
	// Query account balance for the sent denom
	balanceQueryReq := banktypes.NewQueryBalanceRequest(sdk.AccAddress(sender), denom)
	resp, err := accountstd.QueryModule[banktypes.QueryBalanceResponse](ctx, balanceQueryReq)
	if err != nil {
		return nil, err
	}

	return resp.Balance, nil
}

func (bva BaseLockup) checkTokensSendable(ctx context.Context, sender string, amount, lockedCoins sdk.Coins) error {
	// Check if any sent tokens is exceeds lockup account balances
	for _, coin := range amount {
		balance, err := bva.getBalance(ctx, sender, coin.Denom)
		if err != nil {
			return err
		}
		lockedAmt := lockedCoins.AmountOf(coin.Denom)

		// get lockedCoin from that are not bonded for the sent denom
		notBondedLockedCoin, err := bva.GetNotBondedLockedCoin(ctx, sdk.NewCoin(coin.Denom, lockedAmt), coin.Denom)
		if err != nil {
			return err
		}

		spendable, hasNeg := sdk.Coins{*balance}.SafeSub(notBondedLockedCoin)
		if hasNeg {
			return errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds,
				"locked amount exceeds account balance funds: %s > %s", notBondedLockedCoin, balance)
		}

		if _, hasNeg := spendable.SafeSub(coin); hasNeg {
			if len(spendable) == 0 {
				spendable = sdk.Coins{sdk.NewCoin(coin.Denom, math.ZeroInt())}
			}
			return errorsmod.Wrapf(
				sdkerrors.ErrInsufficientFunds,
				"spendable balance %s is smaller than %s",
				spendable, coin,
			)
		}
	}

	return nil
}

// IterateSendEnabledEntries iterates over all the SendEnabled entries.
func (bva BaseLockup) IterateCoinEntries(
	ctx context.Context,
	entries collections.Map[string, math.Int],
	cb func(denom string, value math.Int) (bool, error),
) error {
	err := entries.Walk(ctx, nil, func(key string, value math.Int) (stop bool, err error) {
		return cb(key, value)
	})
	return err
}

// GetNotBondedLockedCoin returns the coin that are not spendable that are not bonded by denom
// for a lockup account. If the coin by the provided denom are not locked, an coin with zero amount is returned.
func (bva BaseLockup) GetNotBondedLockedCoin(ctx context.Context, lockedCoin sdk.Coin, denom string) (sdk.Coin, error) {
	delegatedLockingAmt, err := bva.DelegatedLocking.Get(ctx, denom)
	if err != nil {
		return sdk.Coin{}, err
	}

	x := math.MinInt(lockedCoin.Amount, delegatedLockingAmt)
	lockedAmt := lockedCoin.Amount.Sub(x)

	return sdk.NewCoin(denom, lockedAmt), nil
}

// QueryLockupAccountBaseInfo returns a lockup account's info
func (bva BaseLockup) QueryLockupAccountBaseInfo(ctx context.Context, _ *lockuptypes.QueryLockupAccountInfoRequest) (
	*lockuptypes.QueryLockupAccountInfoResponse, error,
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
	err = bva.IterateCoinEntries(ctx, bva.OriginalLocking, func(key string, value math.Int) (stop bool, err error) {
		originalLocking = append(originalLocking, sdk.NewCoin(key, value))
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	delegatedLocking := sdk.Coins{}
	err = bva.IterateCoinEntries(ctx, bva.DelegatedLocking, func(key string, value math.Int) (stop bool, err error) {
		delegatedLocking = append(delegatedLocking, sdk.NewCoin(key, value))
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	delegatedFree := sdk.Coins{}
	err = bva.IterateCoinEntries(ctx, bva.DelegatedFree, func(key string, value math.Int) (stop bool, err error) {
		delegatedFree = append(delegatedFree, sdk.NewCoin(key, value))
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return &lockuptypes.QueryLockupAccountInfoResponse{
		Owner:            ownerAddress,
		OriginalLocking:  originalLocking,
		DelegatedLocking: delegatedLocking,
		DelegatedFree:    delegatedFree,
		EndTime:          &endTime,
	}, nil
}
