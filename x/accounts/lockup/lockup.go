package vesting

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
	stakingtypes "cosmossdk.io/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	OriginalLockupPrefix  = collections.NewPrefix(0)
	DelegatedFreePrefix   = collections.NewPrefix(1)
	DelegatedLockupPrefix = collections.NewPrefix(2)
	EndTimePrefix         = collections.NewPrefix(3)
	StartTimePrefix       = collections.NewPrefix(4)
	LockupPeriodsPrefix   = collections.NewPrefix(5)
	OwnerPrefix           = collections.NewPrefix(6)
)

var (
	CONTINUOUS_LOCKUP_ACCOUNT = "continuous-lockup-account"
	DELAYED_LOCKUP_ACCOUNT    = "delayed-lockup-account"
	PERIODIC_LOCKUP_ACCOUNT   = "periodic-lockup-account"
	PERMANENT_LOCKUP_ACCOUNT  = "permanent-lockup-account"
)

var (
	MSG_DELEGATE   = "/cosmos.staking.v1beta1.MsgDelegate"
	MSG_UNDELEGATE = "/cosmos.staking.v1beta1.MsgUndelegate"
	MSG_SEND       = "/cosmos.bank.v1beta1.MsgSend"
	MSG_MULTI_SEND = "/cosmos.bank.v1beta1.MsgMultiSend"
)

type getLockCoinsFunc = func(ctx context.Context, time time.Time, denoms ...string) (sdk.Coins, error)

// NewBaseLockup creates a new BaseLockup object.
func NewBaseLockup(d accountstd.Dependencies) *BaseLockup {
	BaseLockup := &BaseLockup{
		Owner:           collections.NewItem(d.SchemaBuilder, OwnerPrefix, "owner", collections.BytesValue),
		OriginalLockup:  collections.NewMap(d.SchemaBuilder, OriginalLockupPrefix, "original_lockup", collections.StringKey, sdk.IntValue),
		DelegatedFree:   collections.NewMap(d.SchemaBuilder, DelegatedFreePrefix, "delegated_free", collections.StringKey, sdk.IntValue),
		DelegatedLockup: collections.NewMap(d.SchemaBuilder, DelegatedLockupPrefix, "delegated_lockup", collections.StringKey, sdk.IntValue),
		addressCodec:    d.AddressCodec,
		headerService:   d.HeaderService,
		EndTime:         collections.NewItem(d.SchemaBuilder, EndTimePrefix, "end_time", collcodec.KeyToValueCodec[time.Time](sdk.TimeKey)),
	}

	return BaseLockup
}

type BaseLockup struct {
	// Owner is the address of the account owner.
	Owner           collections.Item[[]byte]
	OriginalLockup  collections.Map[string, math.Int]
	DelegatedFree   collections.Map[string, math.Int]
	DelegatedLockup collections.Map[string, math.Int]
	addressCodec    address.Codec
	headerService   header.Service
	// lockup end time.
	EndTime collections.Item[time.Time]
}

func (bva *BaseLockup) Init(ctx context.Context, msg *lockuptypes.MsgInitVestingAccount) (
	*lockuptypes.MsgInitVestingAccountResponse, error,
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
	if !funds.Equal(msg.Amount) {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("invalid funding amount, should be equal to lockup amount")
	}

	if err := validateAmount(msg.Amount); err != nil {
		return nil, err
	}

	sortedAmt := msg.Amount.Sort()
	for _, coin := range sortedAmt {
		err = bva.OriginalLockup.Set(ctx, coin.Denom, coin.Amount)
		if err != nil {
			return nil, err
		}
	}

	err = bva.EndTime.Set(ctx, msg.EndTime)
	if err != nil {
		return nil, err
	}

	return &lockuptypes.MsgInitVestingAccountResponse{}, nil
}

// ExecuteMessages handle the execution of codectypes Any messages
// and update the vesting account DelegatedFree and DelegatedVesting
// when delegate or undelegate is trigger. And check for locked coins
// when performing a send message.
func (bva *BaseLockup) ExecuteMessages(
	ctx context.Context, msg *lockuptypes.MsgExecuteMessages, getLockCoinsFunc getLockCoinsFunc,
) (
	*lockuptypes.MsgExecuteMessagesResponse, error,
) {
	owner, err := bva.Owner.Get(ctx)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid owner address: %s", err.Error())
	}
	sender, err := bva.addressCodec.StringToBytes(msg.Sender)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err.Error())
	}
	if !bytes.Equal(owner, sender) {
		return nil, fmt.Errorf("sender is not the owner of this vesting account")
	}
	hs := bva.headerService.GetHeaderInfo(ctx)

	for _, m := range msg.Messages {
		concreteMsg, err := lockuptypes.UnpackAnyRaw(m)
		if err != nil {
			return nil, err
		}

		typeUrl := sdk.MsgTypeURL(concreteMsg)
		switch typeUrl {
		case MSG_DELEGATE:
			msgDelegate, ok := concreteMsg.(*stakingtypes.MsgDelegate)
			if !ok {
				return nil, fmt.Errorf("invalid proto msg for type: %s", typeUrl)
			}
			balance, err := bva.getBalance(ctx, msgDelegate.DelegatorAddress, msgDelegate.Amount.Denom)
			if err != nil {
				return nil, err
			}
			vestingCoins, err := getLockCoinsFunc(ctx, hs.Time, msgDelegate.Amount.Denom)
			if err != nil {
				return nil, err
			}

			err = bva.TrackDelegation(
				ctx,
				sdk.Coins{*balance},
				vestingCoins,
				sdk.Coins{msgDelegate.Amount},
			)
			if err != nil {
				return nil, err
			}
		case MSG_UNDELEGATE:
			msgUndelegate, ok := concreteMsg.(*stakingtypes.MsgUndelegate)
			if !ok {
				return nil, fmt.Errorf("invalid proto msg for type: %s", typeUrl)
			}

			err = bva.TrackUndelegation(ctx, sdk.Coins{msgUndelegate.Amount})
			if err != nil {
				return nil, err
			}
		case MSG_SEND:
			msgSend, ok := concreteMsg.(*banktypes.MsgSend)
			if !ok {
				return nil, fmt.Errorf("invalid proto msg for type: %s", typeUrl)
			}
			sender := msgSend.FromAddress
			amount := msgSend.Amount

			vestingCoins, err := getLockCoinsFunc(ctx, hs.Time, amount.Denoms()...)
			if err != nil {
				return nil, err
			}

			err = bva.checkTokensSendable(ctx, sender, amount, vestingCoins)
			if err != nil {
				return nil, err
			}
		case MSG_MULTI_SEND:
			msgMultiSend, ok := concreteMsg.(*banktypes.MsgMultiSend)
			if !ok {
				return nil, fmt.Errorf("invalid proto msg for type: %s", typeUrl)
			}
			sender := msgMultiSend.Inputs[0].Address
			amount := msgMultiSend.Inputs[0].Coins

			vestingCoins, err := getLockCoinsFunc(ctx, hs.Time, amount.Denoms()...)
			if err != nil {
				return nil, err
			}

			err = bva.checkTokensSendable(ctx, sender, amount, vestingCoins)
			if err != nil {
				return nil, err
			}
		}
	}

	// execute messages
	responses, err := accountstd.ExecModuleAnys(ctx, msg.Messages)
	if err != nil {
		return nil, err
	}

	return &lockuptypes.MsgExecuteMessagesResponse{Responses: responses}, nil
}

// TrackDelegation tracks a delegation amount for any given lockup account type
// given the amount of coins currently being locked and the current account balance
// of the delegation denominations.
//
// CONTRACT: The account's coins, delegation coins, locked coins, and delegated
// lockup coins must be sorted.
func (bva *BaseLockup) TrackDelegation(
	ctx context.Context, balance, vestingCoins, amount sdk.Coins,
) error {
	for _, coin := range amount {
		baseAmt := balance.AmountOf(coin.Denom)
		vestingAmt := vestingCoins.AmountOf(coin.Denom)
		delVestingAmt, err := bva.DelegatedLockup.Get(ctx, coin.Denom)
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
		x := math.MinInt(math.MaxInt(vestingAmt.Sub(delVestingAmt), math.ZeroInt()), coin.Amount)
		y := coin.Amount.Sub(x)

		delVestingCoin := sdk.NewCoin(coin.Denom, delVestingAmt)
		delFreeCoin := sdk.NewCoin(coin.Denom, delFreeAmt)
		if !x.IsZero() {
			xCoin := sdk.NewCoin(coin.Denom, x)
			newDelVesting := delVestingCoin.Add(xCoin)
			err = bva.DelegatedLockup.Set(ctx, newDelVesting.Denom, newDelVesting.Amount)
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
// values by which delegated vesting and delegated vesting need to decrease and
// by which amount the base coins need to increase.
//
// NOTE: The undelegation (bond refund) amount may exceed the delegated
// vesting (bond) amount due to the way undelegation truncates the bond refund,
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
		delVestingAmt, err := bva.DelegatedLockup.Get(ctx, coin.Denom)
		if err != nil {
			return err
		}

		// compute x and y per the specification, where:
		// X := min(DF, D)
		// Y := min(DV, D - X)
		x := math.MinInt(delFreeAmt, coin.Amount)
		y := math.MinInt(delVestingAmt, coin.Amount.Sub(x))

		delVestingCoin := sdk.NewCoin(coin.Denom, delVestingAmt)
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
			newDelVesting := delVestingCoin.Sub(yCoin)
			err = bva.DelegatedLockup.Set(ctx, newDelVesting.Denom, newDelVesting.Amount)
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

func (bva BaseLockup) checkTokensSendable(ctx context.Context, sender string, amount, vestingCoins sdk.Coins) error {
	// Check if any sent tokens is exceeds vesting account balances
	for _, coin := range amount {
		balance, err := bva.getBalance(ctx, sender, coin.Denom)
		if err != nil {
			return err
		}
		vestingAmt := vestingCoins.AmountOf(coin.Denom)

		// get lockedCoin for the sent denom
		locked, err := bva.LockedCoinFromVesting(ctx, sdk.NewCoin(coin.Denom, vestingAmt), coin.Denom)
		if err != nil {
			return err
		}

		spendable, hasNeg := sdk.Coins{*balance}.SafeSub(locked)
		if hasNeg {
			return errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds,
				"locked amount exceeds account balance funds: %s > %s", locked, balance)
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

// LockedCoinsFromVesting returns the coin that are not spendable by denom (i.e. locked)
// for a vesting account given the current vesting coin. If the coin by the provided denom
// are not locked, an coin with zero amount is returned.
func (bva BaseLockup) LockedCoinFromVesting(ctx context.Context, vestingCoin sdk.Coin, denom string) (sdk.Coin, error) {
	delegatedVestingAmt, err := bva.DelegatedLockup.Get(ctx, denom)
	if err != nil {
		return sdk.Coin{}, err
	}
	vestingAmt := vestingCoin.Amount

	x := math.MinInt(vestingAmt, delegatedVestingAmt)
	lockedAmt := vestingAmt.Sub(x)

	lockedCoin := sdk.NewCoin(denom, lockedAmt)

	return lockedCoin, nil
}

// QueryVestingAccountInfo returns a vesting account's info
func (bva BaseLockup) QueryVestingAccountBaseInfo(ctx context.Context, _ *lockuptypes.QueryVestingAccountInfoRequest) (
	*lockuptypes.QueryVestingAccountInfoResponse, error,
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

	originalVesting := sdk.Coins{}
	err = bva.IterateCoinEntries(ctx, bva.OriginalLockup, func(key string, value math.Int) (stop bool, err error) {
		originalVesting = append(originalVesting, sdk.NewCoin(key, value))
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	delegatedVesting := sdk.Coins{}
	err = bva.IterateCoinEntries(ctx, bva.DelegatedLockup, func(key string, value math.Int) (stop bool, err error) {
		delegatedVesting = append(delegatedVesting, sdk.NewCoin(key, value))
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

	return &lockuptypes.QueryVestingAccountInfoResponse{
		Owner:            ownerAddress,
		OriginalVesting:  originalVesting,
		DelegatedVesting: delegatedVesting,
		DelegatedFree:    delegatedFree,
		EndTime:          &endTime,
	}, nil
}
