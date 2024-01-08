package vesting

import (
	"context"
	"time"

	"cosmossdk.io/x/accounts/accountstd"
	vestingtypes "cosmossdk.io/x/accounts/vesting/types/v1"
	banktypes "cosmossdk.io/x/bank/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Compile-time type assertions
var (
	_ accountstd.Interface = (*DelayedVestingAccount)(nil)
)

// Delayed Vesting Account

// NewDelayedVestingAccount creates a new DelayedVestingAccount object.
func NewDelayedVestingAccount(d accountstd.Dependencies) (*DelayedVestingAccount, error) {
	baseVestingAccount := BaseVestingAccount{
		OriginalVesting:  sdk.NewCoins(),
		DelegatedFree:    sdk.NewCoins(),
		DelegatedVesting: sdk.NewCoins(),
		AddressCodec:     d.AddressCodec,
		EndTime:          0,
	}

	return &DelayedVestingAccount{
		&baseVestingAccount,
	}, nil
}

type DelayedVestingAccount struct {
	*BaseVestingAccount
}

// --------------- Init -----------------

func (dva DelayedVestingAccount) Init(ctx context.Context, msg *vestingtypes.MsgInitVestingAccount) (*vestingtypes.MsgInitVestingAccountResponse, error) {
	to := accountstd.Whoami(ctx)
	if to == nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("Cannot find account address from context")
	}

	toAddress, err := dva.AddressCodec.BytesToString(to)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid 'to' address: %s", err)
	}

	if err := validateAmount(msg.Amount); err != nil {
		return nil, err
	}

	if msg.EndTime <= 0 {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("invalid end time")
	}

	dva.OriginalVesting = msg.Amount.Sort()
	dva.DelegatedFree = sdk.NewCoins()
	dva.DelegatedVesting = sdk.NewCoins()
	dva.EndTime = msg.EndTime

	// Send token to new vesting account
	sendMsg := banktypes.NewMsgSend(msg.FromAddress, toAddress, msg.Amount)
	anyMsg, err := codectypes.NewAnyWithValue(sendMsg)
	if err != nil {
		return nil, err
	}

	if _, err = accountstd.ExecModuleAnys(ctx, []*codectypes.Any{anyMsg}); err != nil {
		return nil, err
	}

	// Validate the newly init account
	err = dva.BaseVestingAccount.Validate()
	if err != nil {
		return nil, err
	}

	return &vestingtypes.MsgInitVestingAccountResponse{}, nil
}

// --------------- execute -----------------

// LockedCoins returns the set of coins that are not spendable (i.e. locked),
// defined as the vesting coins that are not delegated.
func (dva DelayedVestingAccount) LockedCoins(blockTime time.Time) sdk.Coins {
	return dva.BaseVestingAccount.LockedCoinsFromVesting(dva.GetVestingCoins(blockTime))
}

// TrackDelegation tracks a desired delegation amount by setting the appropriate
// values for the amount of delegated vesting, delegated free, and reducing the
// overall amount of base coins.
func (dva *DelayedVestingAccount) TrackDelegation(blockTime time.Time, balance, amount sdk.Coins) {
	dva.BaseVestingAccount.TrackDelegation(balance, dva.GetVestingCoins(blockTime), amount)
}

// --------------- Query -----------------

// GetVestedCoins returns the total number of vested coins. If no coins are vested,
// nil is returned.
func (dva DelayedVestingAccount) GetVestedCoins(blockTime time.Time) sdk.Coins {
	if blockTime.Unix() >= dva.EndTime {
		return dva.OriginalVesting
	}

	return nil
}

// GetVestingCoins returns the total number of vesting coins. If no coins are
// vesting, nil is returned.
func (dva DelayedVestingAccount) GetVestingCoins(blockTime time.Time) sdk.Coins {
	return dva.OriginalVesting.Sub(dva.GetVestedCoins(blockTime)...)
}

func (dva DelayedVestingAccount) QueryVestedCoins(ctx context.Context, msg *vestingtypes.QueryVestedCoinsRequest) (
	*vestingtypes.QueryVestedCoinsResponse, error,
) {
	originalContext := accountstd.OriginalContext(ctx)
	sdkctx := sdk.UnwrapSDKContext(originalContext)
	vestedCoins := dva.GetVestedCoins(sdkctx.HeaderInfo().Time)

	return &vestingtypes.QueryVestedCoinsResponse{
		VestedVesting: vestedCoins,
	}, nil
}

func (dva DelayedVestingAccount) QueryVestingCoins(ctx context.Context, msg *vestingtypes.QueryVestingCoinsRequest) (
	*vestingtypes.QueryVestingCoinsResponse, error,
) {
	originalContext := accountstd.OriginalContext(ctx)
	sdkctx := sdk.UnwrapSDKContext(originalContext)
	vestingCoins := dva.GetVestingCoins(sdkctx.BlockHeader().Time)

	return &vestingtypes.QueryVestingCoinsResponse{
		VestingCoins: vestingCoins,
	}, nil
}

// Implement smart account interface
func (dva DelayedVestingAccount) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, dva.Init)
}

func (dva DelayedVestingAccount) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
}

func (dva DelayedVestingAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, dva.QueryVestedCoins)
	accountstd.RegisterQueryHandler(builder, dva.QueryVestingCoins)
	dva.BaseVestingAccount.RegisterQueryHandlers(builder)
}
