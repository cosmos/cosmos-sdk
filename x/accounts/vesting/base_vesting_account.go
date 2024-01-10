package vesting

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	impl "cosmossdk.io/x/accounts/internal/implementation"
	vestingtypes "cosmossdk.io/x/accounts/vesting/types/v1"
	vestingtypesv1 "cosmossdk.io/x/accounts/vesting/types/v1"
	banktypes "cosmossdk.io/x/bank/types"
	stakingtypes "cosmossdk.io/x/staking/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
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
)

// Base Vesting Account
var _ accountstd.Interface = (*BaseVestingAccount)(nil)

type getVestingFunc = func(ctx context.Context, time time.Time) (sdk.Coins, error)

// NewBaseVestingAccount creates a new BaseVestingAccount object.
func NewBaseVestingAccount(d accountstd.Dependencies) (*BaseVestingAccount, error) {
	baseVestingAccount := &BaseVestingAccount{
		OriginalVesting:  collections.NewMap(d.SchemaBuilder, OriginalVestingPrefix, "original_vesting", collections.StringKey, sdk.IntValue),
		DelegatedFree:    collections.NewMap(d.SchemaBuilder, DelegatedFreePrefix, "delegated_free", collections.StringKey, sdk.IntValue),
		DelegatedVesting: collections.NewMap(d.SchemaBuilder, DelegatedVestingPrefix, "delegated_vesting", collections.StringKey, sdk.IntValue),
		AddressCodec:     d.AddressCodec,
		EndTime:          collections.NewItem(d.SchemaBuilder, EndTimePrefix, "end_time", sdk.IntValue),
	}

	return baseVestingAccount, nil
}

type BaseVestingAccount struct {
	OriginalVesting  collections.Map[string, math.Int]
	DelegatedFree    collections.Map[string, math.Int]
	DelegatedVesting collections.Map[string, math.Int]
	AddressCodec     address.Codec
	// Vesting end time, as unix timestamp (in seconds).
	EndTime collections.Item[math.Int]
}

func (bva *BaseVestingAccount) Init(ctx context.Context, msg *vestingtypes.MsgInitVestingAccount) (
	*vestingtypes.MsgInitVestingAccountResponse, error,
) {
	to := accountstd.Whoami(ctx)
	if to == nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("Cannot find account address from context")
	}

	toAddress, err := bva.AddressCodec.BytesToString(to)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid 'to' address: %s", err)
	}

	if err := validateAmount(msg.Amount); err != nil {
		return nil, err
	}

	sortedAmt := msg.Amount.Sort()
	for _, coin := range sortedAmt {
		bva.OriginalVesting.Set(ctx, coin.Denom, coin.Amount)
	}

	bva.EndTime.Set(ctx, math.NewInt(msg.EndTime))

	// Send token to new vesting account
	sendMsg := banktypes.NewMsgSend(msg.FromAddress, toAddress, msg.Amount)
	anyMsg, err := codectypes.NewAnyWithValue(sendMsg)
	if err != nil {
		return nil, err
	}

	if _, err = accountstd.ExecModuleAnys(ctx, []*codectypes.Any{anyMsg}); err != nil {
		return nil, err
	}

	return &vestingtypes.MsgInitVestingAccountResponse{}, nil
}

// --------------- execute -----------------

// TrackDelegation tracks a delegation amount for any given vesting account type
// given the amount of coins currently vesting and the current account balance
// of the delegation denominations.
//
// CONTRACT: The account's coins, delegation coins, vesting coins, and delegated
// vesting coins must be sorted.
func (bva *BaseVestingAccount) TrackDelegation(
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

		// Panic if the delegation amount is zero or if the base coins does not
		// exceed the desired delegation amount.
		if coin.Amount.IsZero() || baseAmt.LT(coin.Amount) {
			panic("delegation attempt with zero coins or insufficient funds")
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
			bva.DelegatedVesting.Set(ctx, newDelVesting.Denom, newDelVesting.Amount)
		}

		if !y.IsZero() {
			yCoin := sdk.NewCoin(coin.Denom, y)
			newDelFree := delFreeCoin.Add(yCoin)
			bva.DelegatedFree.Set(ctx, newDelFree.Denom, newDelFree.Amount)
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
func (bva *BaseVestingAccount) TrackUndelegation(ctx context.Context, amount sdk.Coins) error {
	for _, coin := range amount {
		// panic if the undelegation amount is zero
		if coin.Amount.IsZero() {
			panic("undelegation attempt with zero coins")
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
			bva.DelegatedVesting.Set(ctx, newDelFree.Denom, newDelFree.Amount)
		}

		if !y.IsZero() {
			yCoin := sdk.NewCoin(coin.Denom, y)
			newDelVesting := delVestingCoin.Sub(yCoin)
			bva.DelegatedVesting.Set(ctx, newDelVesting.Denom, newDelVesting.Amount)
		}
	}

	return nil
}

// ExecuteMessages handle the execution of codectypes Any messages
// and update the vesting account DelegatedFree and DelegatedVesting
// when delegate or undelegate is trigger.
func (bva *BaseVestingAccount) ExecuteMessages(
	ctx context.Context, msg *vestingtypesv1.MsgExecuteMessages, getVestingFunc getVestingFunc,
) (
	*vestingtypesv1.MsgExecuteMessagesResponse, error,
) {
	originalContext := accountstd.OriginalContext(ctx)
	sdkctx := sdk.UnwrapSDKContext(originalContext)

	for _, m := range msg.ExecutionMessages {
		concreteMessage, err := impl.UnpackAnyRaw(m)
		if err != nil {
			return nil, err
		}

		typeUrl := codectypes.MsgTypeURL(concreteMessage)
		switch typeUrl {
		case "/cosmos.staking.v1beta1.MsgDelegate":
			msgDelegate, ok := concreteMessage.(*stakingtypes.MsgDelegate)
			if !ok {
				return nil, fmt.Errorf("Invalid proto msg for type: %s", typeUrl)
			}

			// Query account balance for the delegated denom
			balanceQueryReq := banktypes.NewQueryBalanceRequest(sdk.AccAddress(msgDelegate.DelegatorAddress), msgDelegate.Amount.Denom)
			resp, err := accountstd.QueryModule[banktypes.QueryBalanceResponse](ctx, balanceQueryReq)
			if err != nil {
				return nil, err
			}
			vestingCoins, err := getVestingFunc(ctx, sdkctx.BlockHeader().Time)
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
			msgUndelegate, ok := concreteMessage.(*stakingtypes.MsgUndelegate)
			if !ok {
				return nil, fmt.Errorf("Invalid proto msg for type: %s", typeUrl)
			}

			err := bva.TrackUndelegation(ctx, sdk.Coins{msgUndelegate.Amount})
			if err != nil {
				return nil, err
			}
		case "/cosmos.bank.v1beta1.MsgSend", "/cosmos.bank.v1beta1.MsgMultiSend":
			var sender string
			var amount sdk.Coins
			if typeUrl == "/cosmos.bank.v1beta1.MsgSend" {
				msgSend, ok := concreteMessage.(*banktypes.MsgSend)
				if !ok {
					return nil, fmt.Errorf("Invalid proto msg for type: %s", typeUrl)
				}
				sender = msgSend.FromAddress
				amount = msgSend.Amount
			} else {
				msgMultiSend, ok := concreteMessage.(*banktypes.MsgMultiSend)
				if !ok {
					return nil, fmt.Errorf("Invalid proto msg for type: %s", typeUrl)
				}
				sender = msgMultiSend.Inputs[0].Address
				amount = msgMultiSend.Inputs[0].Coins
			}

			vestingCoins, err := getVestingFunc(ctx, sdkctx.BlockHeader().Time)
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
		default:
			sdkctx.Logger().Info("Non special case continue the execution")
		}
	}

	// execute messages
	responses, err := accountstd.ExecModuleAnys(ctx, msg.ExecutionMessages)
	if err != nil {
		return nil, err
	}

	return &vestingtypesv1.MsgExecuteMessagesResponse{ExecutionMessagesResponse: responses}, nil
}

// --------------- Query -----------------

// IterateSendEnabledEntries iterates over all the SendEnabled entries.
func (bva BaseVestingAccount) IterateEntries(
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
func (bva BaseVestingAccount) LockedCoinsFromVesting(ctx context.Context, vestingCoins sdk.Coins) sdk.Coins {
	var delegatedVestingCoins sdk.Coins
	bva.IterateEntries(ctx, bva.DelegatedVesting, func(key string, value math.Int) (stop bool) {
		delegatedVestingCoins = append(delegatedVestingCoins, sdk.NewCoin(key, value))
		return false
	})

	lockedCoins := vestingCoins.Sub(vestingCoins.Min(delegatedVestingCoins)...)
	if lockedCoins == nil {
		return sdk.Coins{}
	}
	return lockedCoins
}

// QueryOriginalVesting returns a vesting account's original vesting amount
func (bva BaseVestingAccount) QueryOriginalVesting(ctx context.Context, _ *vestingtypesv1.QueryOriginalVestingRequest) (
	*vestingtypesv1.QueryOriginalVestingResponse, error,
) {
	var originalVesting sdk.Coins
	bva.IterateEntries(ctx, bva.OriginalVesting, func(key string, value math.Int) (stop bool) {
		originalVesting = append(originalVesting, sdk.NewCoin(key, value))
		return false
	})
	return &vestingtypesv1.QueryOriginalVestingResponse{
		OriginalVesting: originalVesting,
	}, nil
}

// QueryDelegatedFree returns a vesting account's delegation amount that is not
// vesting.
func (bva BaseVestingAccount) QueryDelegatedFree(ctx context.Context, _ *vestingtypesv1.QueryDelegatedFreeRequest) (
	*vestingtypesv1.QueryDelegatedFreeResponse, error,
) {
	var delegatedFree sdk.Coins
	bva.IterateEntries(ctx, bva.DelegatedFree, func(key string, value math.Int) (stop bool) {
		delegatedFree = append(delegatedFree, sdk.NewCoin(key, value))
		return false
	})
	return &vestingtypesv1.QueryDelegatedFreeResponse{
		DelegatedFree: delegatedFree,
	}, nil
}

// QueryDelegatedVesting returns a vesting account's delegation amount that is
// still vesting.
func (bva BaseVestingAccount) QueryDelegatedVesting(ctx context.Context, _ *vestingtypesv1.QueryDelegatedVestingRequest) (
	*vestingtypesv1.QueryDelegatedVestingResponse, error,
) {
	var delegatedVesting sdk.Coins
	bva.IterateEntries(ctx, bva.DelegatedVesting, func(key string, value math.Int) (stop bool) {
		delegatedVesting = append(delegatedVesting, sdk.NewCoin(key, value))
		return false
	})
	return &vestingtypesv1.QueryDelegatedVestingResponse{
		DelegatedVesting: delegatedVesting,
	}, nil
}

// QueryEndTime returns a vesting account's end time
func (bva BaseVestingAccount) QueryEndTime(ctx context.Context, _ *vestingtypesv1.QueryEndTimeRequest) (
	*vestingtypesv1.QueryEndTimeResponse, error,
) {
	endTime, err := bva.EndTime.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &vestingtypesv1.QueryEndTimeResponse{
		EndTime: endTime.Int64(),
	}, nil
}

// Only for implementing account interface, base vesting account
// served as a base for other types of vesting account only
// and should not be initialize as a stand alone vesting account type.
func (bva BaseVestingAccount) RegisterInitHandler(builder *accountstd.InitBuilder) {
}

func (bva BaseVestingAccount) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
}

func (bva BaseVestingAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, bva.QueryOriginalVesting)
	accountstd.RegisterQueryHandler(builder, bva.QueryDelegatedFree)
	accountstd.RegisterQueryHandler(builder, bva.QueryDelegatedVesting)
	accountstd.RegisterQueryHandler(builder, bva.QueryEndTime)
}
