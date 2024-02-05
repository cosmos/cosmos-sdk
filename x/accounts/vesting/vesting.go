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
	account_abstractionv1 "cosmossdk.io/x/accounts/interfaces/account_abstraction/v1"
	vestingtypes "cosmossdk.io/x/accounts/vesting/types"
	banktypes "cosmossdk.io/x/bank/types"
	stakingtypes "cosmossdk.io/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	OriginalVestingPrefix  = collections.NewPrefix(0)
	DelegatedFreePrefix    = collections.NewPrefix(1)
	DelegatedVestingPrefix = collections.NewPrefix(2)
	EndTimePrefix          = collections.NewPrefix(3)
	StartTimePrefix        = collections.NewPrefix(4)
	VestingPeriodsPrefix   = collections.NewPrefix(5)
	OwnerPrefix            = collections.NewPrefix(6)
)

var (
	CONTINUOUS_VESTING_ACCOUNT = "continuous-vesting-account"
	DELAYED_VESTING_ACCOUNT    = "delayed-vesting-account"
	PERIODIC_VESTING_ACCOUNT   = "periodic-vesting-account"
	PERMANENT_VESTING_ACCOUNT  = "permanent-vesting-account"
)

type getVestingFunc = func(ctx context.Context, time time.Time) (sdk.Coins, error)

// NewBaseVesting creates a new BaseVesting object.
func NewBaseVesting(d accountstd.Dependencies) *BaseVesting {
	baseVesting := &BaseVesting{
		Owner:            collections.NewItem(d.SchemaBuilder, OwnerPrefix, "owner", collections.BytesValue),
		OriginalVesting:  collections.NewMap(d.SchemaBuilder, OriginalVestingPrefix, "original_vesting", collections.StringKey, sdk.IntValue),
		DelegatedFree:    collections.NewMap(d.SchemaBuilder, DelegatedFreePrefix, "delegated_free", collections.StringKey, sdk.IntValue),
		DelegatedVesting: collections.NewMap(d.SchemaBuilder, DelegatedVestingPrefix, "delegated_vesting", collections.StringKey, sdk.IntValue),
		addressCodec:     d.AddressCodec,
		headerService:    d.HeaderService,
		EndTime:          collections.NewItem(d.SchemaBuilder, EndTimePrefix, "end_time", collcodec.KeyToValueCodec[time.Time](sdk.TimeKey)),
	}

	return baseVesting
}

type BaseVesting struct {
	// Owner is the address of the account owner.
	Owner            collections.Item[[]byte]
	OriginalVesting  collections.Map[string, math.Int]
	DelegatedFree    collections.Map[string, math.Int]
	DelegatedVesting collections.Map[string, math.Int]
	addressCodec     address.Codec
	headerService    header.Service
	// Vesting end time, as unix timestamp (in seconds).
	EndTime collections.Item[time.Time]
}

func (bva *BaseVesting) Init(ctx context.Context, msg *vestingtypes.MsgInitVestingAccount) (
	*vestingtypes.MsgInitVestingAccountResponse, error,
) {
	sender := accountstd.Sender(ctx)
	to := accountstd.Whoami(ctx)

	toAddress, err := bva.addressCodec.BytesToString(to)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid 'to' address: %s", err)
	}
	fromAddress, err := bva.addressCodec.BytesToString(sender)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid 'sender' address: %s", err)
	}
	owner, err := bva.addressCodec.StringToBytes(msg.Owner)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid 'owner' address: %s", err)
	}
	err = bva.Owner.Set(ctx, owner)
	if err != nil {
		return nil, err
	}

	if err := validateAmount(msg.Amount); err != nil {
		return nil, err
	}

	sortedAmt := msg.Amount.Sort()
	for _, coin := range sortedAmt {
		err = bva.OriginalVesting.Set(ctx, coin.Denom, coin.Amount)
		if err != nil {
			return nil, err
		}
	}

	err = bva.EndTime.Set(ctx, msg.EndTime)
	if err != nil {
		return nil, err
	}

	// Send token to new vesting account
	sendMsg := banktypes.NewMsgSend(fromAddress, toAddress, msg.Amount)
	if _, err = accountstd.ExecModule[banktypes.MsgSendResponse](ctx, sendMsg); err != nil {
		return nil, err
	}

	return &vestingtypes.MsgInitVestingAccountResponse{}, nil
}

// TrackDelegation tracks a delegation amount for any given vesting account type
// given the amount of coins currently vesting and the current account balance
// of the delegation denominations.
//
// CONTRACT: The account's coins, delegation coins, vesting coins, and delegated
// vesting coins must be sorted.
func (bva *BaseVesting) TrackDelegation(
	ctx context.Context, balance, vestingCoins, amount sdk.Coins,
) error {
	for _, coin := range amount {
		baseAmt := balance.AmountOf(coin.Denom)
		vestingAmt := vestingCoins.AmountOf(coin.Denom)
		delVestingAmt, err := bva.DelegatedVesting.Get(ctx, coin.Denom)
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
			err = bva.DelegatedVesting.Set(ctx, newDelVesting.Denom, newDelVesting.Amount)
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
func (bva *BaseVesting) TrackUndelegation(ctx context.Context, amount sdk.Coins) error {
	for _, coin := range amount {
		// return error if the undelegation amount is zero
		if coin.Amount.IsZero() {
			return sdkerrors.ErrInvalidCoins.Wrap("undelegation attempt with zero coins")
		}
		delFreeAmt, err := bva.DelegatedFree.Get(ctx, coin.Denom)
		if err != nil {
			return err
		}
		delVestingAmt, err := bva.DelegatedVesting.Get(ctx, coin.Denom)
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
			err = bva.DelegatedVesting.Set(ctx, newDelFree.Denom, newDelFree.Amount)
			if err != nil {
				return err
			}
		}

		if !y.IsZero() {
			yCoin := sdk.NewCoin(coin.Denom, y)
			newDelVesting := delVestingCoin.Sub(yCoin)
			err = bva.DelegatedVesting.Set(ctx, newDelVesting.Denom, newDelVesting.Amount)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ExecuteMessages handle the execution of codectypes Any messages
// and update the vesting account DelegatedFree and DelegatedVesting
// when delegate or undelegate is trigger. And check for locked coins
// when performing a send message.
func (bva *BaseVesting) ExecuteMessages(
	ctx context.Context, msg *account_abstractionv1.MsgExecute, getVestingFunc getVestingFunc,
) (
	*account_abstractionv1.MsgExecuteResponse, error,
) {
	// we always want to ensure that this is called by the x/accounts module, it's the only trusted entrypoint.
	// if we do not do this check then someone could call this method directly and bypass the authentication.
	if !accountstd.SenderIsAccountsModule(ctx) {
		return nil, fmt.Errorf("sender is not the x/accounts module")
	}
	hs := bva.headerService.GetHeaderInfo(ctx)

	for _, m := range msg.ExecutionMessages {
		typeUrl := m.TypeUrl
		switch typeUrl {
		case "/cosmos.staking.v1beta1.MsgDelegate":
			msgDelegate, err := accountstd.UnpackAny[stakingtypes.MsgDelegate](m)
			if err != nil {
				return nil, fmt.Errorf("invalid proto msg for type: %s", typeUrl)
			}

			// Query account balance for the delegated denom
			balanceQueryReq := banktypes.NewQueryBalanceRequest(sdk.AccAddress(msgDelegate.DelegatorAddress), msgDelegate.Amount.Denom)
			resp, err := accountstd.QueryModule[banktypes.QueryBalanceResponse](ctx, balanceQueryReq)
			if err != nil {
				return nil, err
			}
			vestingCoins, err := getVestingFunc(ctx, hs.Time)
			if err != nil {
				return nil, err
			}

			err = bva.TrackDelegation(
				ctx,
				sdk.Coins{*resp.Balance},
				vestingCoins,
				sdk.Coins{msgDelegate.Amount},
			)
			if err != nil {
				return nil, err
			}
		case "/cosmos.staking.v1beta1.MsgUndelegate":
			msgUndelegate, err := accountstd.UnpackAny[stakingtypes.MsgUndelegate](m)
			if err != nil {
				return nil, fmt.Errorf("invalid proto msg for type: %s", typeUrl)
			}

			err = bva.TrackUndelegation(ctx, sdk.Coins{msgUndelegate.Amount})
			if err != nil {
				return nil, err
			}
		case "/cosmos.bank.v1beta1.MsgSend", "/cosmos.bank.v1beta1.MsgMultiSend":
			var sender string
			var amount sdk.Coins
			if typeUrl == "/cosmos.bank.v1beta1.MsgSend" {
				msgSend, err := accountstd.UnpackAny[banktypes.MsgSend](m)
				if err != nil {
					return nil, fmt.Errorf("invalid proto msg for type: %s", typeUrl)
				}
				sender = msgSend.FromAddress
				amount = msgSend.Amount
			} else {
				msgMultiSend, err := accountstd.UnpackAny[banktypes.MsgMultiSend](m)
				if err != nil {
					return nil, fmt.Errorf("invalid proto msg for type: %s", typeUrl)
				}
				sender = msgMultiSend.Inputs[0].Address
				amount = msgMultiSend.Inputs[0].Coins
			}

			vestingCoins, err := getVestingFunc(ctx, hs.Time)
			if err != nil {
				return nil, err
			}

			// Get locked token
			lockedCoins := bva.LockedCoinsFromVesting(ctx, vestingCoins)

			// Check if any sent tokens is exceeds vesting account balances
			for _, coin := range amount {
				// Query account balance for the sent denom
				balanceQueryReq := banktypes.NewQueryBalanceRequest(sdk.AccAddress(sender), coin.Denom)
				resp, err := accountstd.QueryModule[banktypes.QueryBalanceResponse](ctx, balanceQueryReq)
				if err != nil {
					return nil, err
				}
				balance := resp.Balance
				locked := sdk.NewCoin(coin.Denom, lockedCoins.AmountOf(coin.Denom))

				spendable, hasNeg := sdk.Coins{*balance}.SafeSub(locked)
				if hasNeg {
					return nil, errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds,
						"locked amount exceeds account balance funds: %s > %s", locked, balance)
				}

				if _, hasNeg := spendable.SafeSub(coin); hasNeg {
					if len(spendable) == 0 {
						spendable = sdk.Coins{sdk.NewCoin(coin.Denom, math.ZeroInt())}
					}
					return nil, errorsmod.Wrapf(
						sdkerrors.ErrInsufficientFunds,
						"spendable balance %s is smaller than %s",
						spendable, coin,
					)
				}
			}
		}
	}

	// execute messages
	responses, err := accountstd.ExecModuleAnys(ctx, msg.ExecutionMessages)
	if err != nil {
		return nil, err
	}

	return &account_abstractionv1.MsgExecuteResponse{ExecutionMessagesResponse: responses}, nil
}

// Check the sender of the bundle is the owner of the vesting account
func (bva BaseVesting) Authenticate(ctx context.Context, msg *account_abstractionv1.MsgAuthenticate) (*account_abstractionv1.MsgAuthenticateResponse, error) {
	owner, err := bva.Owner.Get(ctx)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid owner address: %s", err.Error())
	}
	bundler, err := bva.addressCodec.StringToBytes(msg.Bundler)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid bundler address: %s", err.Error())
	}
	if !bytes.Equal(owner, bundler) {
		return nil, fmt.Errorf("bundler is not the owner of this vesting account")
	}

	// prevent bypass vesting account check for send, delegate and undelegate action
	// when execute messages by pass it in bundler. Since vesting doesn't have a handler
	// for payBundler message, this could lead to accounts keeper executes the messages
	// directly without going through vesting account ExecuteMessages handler.
	if len(msg.UserOperation.BundlerPaymentMessages) > 0 {
		return nil, fmt.Errorf("no payment messages should be included for vesting account")
	}

	return &account_abstractionv1.MsgAuthenticateResponse{}, nil
}

// IterateSendEnabledEntries iterates over all the SendEnabled entries.
func (bva BaseVesting) IterateCoinEntries(
	ctx context.Context,
	entries collections.Map[string, math.Int],
	cb func(denom string, value math.Int) bool,
) {
	err := entries.Walk(ctx, nil, func(key string, value math.Int) (stop bool, err error) {
		return cb(key, value), nil
	})
	if err != nil {
		panic(err)
	}
}

// LockedCoinsFromVesting returns all the coins that are not spendable (i.e. locked)
// for a vesting account given the current vesting coins. If no coins are locked,
// an empty slice of Coins is returned.
//
// CONTRACT: Delegated vesting coins and vestingCoins must be sorted.
func (bva BaseVesting) LockedCoinsFromVesting(ctx context.Context, vestingCoins sdk.Coins) sdk.Coins {
	delegatedVestingCoins := sdk.Coins{}
	bva.IterateCoinEntries(ctx, bva.DelegatedVesting, func(key string, value math.Int) (stop bool) {
		delegatedVestingCoins = append(delegatedVestingCoins, sdk.NewCoin(key, value))
		return false
	})

	lockedCoins := vestingCoins.Sub(vestingCoins.Min(delegatedVestingCoins)...)
	if lockedCoins == nil {
		return sdk.Coins{}
	}
	return lockedCoins
}

// QueryVestingAccountInfo returns a vesting account's info
func (bva BaseVesting) QueryVestingAccountBaseInfo(ctx context.Context, _ *vestingtypes.QueryVestingAccountInfoRequest) (
	*vestingtypes.QueryVestingAccountInfoResponse, error,
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
	bva.IterateCoinEntries(ctx, bva.OriginalVesting, func(key string, value math.Int) (stop bool) {
		originalVesting = append(originalVesting, sdk.NewCoin(key, value))
		return false
	})
	delegatedVesting := sdk.Coins{}
	bva.IterateCoinEntries(ctx, bva.DelegatedVesting, func(key string, value math.Int) (stop bool) {
		delegatedVesting = append(delegatedVesting, sdk.NewCoin(key, value))
		return false
	})
	delegatedFree := sdk.Coins{}
	bva.IterateCoinEntries(ctx, bva.DelegatedFree, func(key string, value math.Int) (stop bool) {
		delegatedFree = append(delegatedFree, sdk.NewCoin(key, value))
		return false
	})
	return &vestingtypes.QueryVestingAccountInfoResponse{
		Owner:            ownerAddress,
		OriginalVesting:  originalVesting,
		DelegatedVesting: delegatedVesting,
		DelegatedFree:    delegatedFree,
		EndTime:          &endTime,
	}, nil
}

func (bva BaseVesting) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, bva.Authenticate)
}
