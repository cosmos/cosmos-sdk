package vesting

import (
	"context"
	"errors"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	vestingtypes "cosmossdk.io/x/accounts/vesting/types/v1"
	banktypes "cosmossdk.io/x/bank/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Compile-time type assertions
var (
	_ accountstd.Interface = (*ContinuousVestingAccount)(nil)
)

// Continuos Vesting Account

// NewContinuousVestingAccount creates a new ContinuousVestingAccount object.
func NewContinuousVestingAccount(d accountstd.Dependencies) (*ContinuousVestingAccount, error) {
	baseVestingAccount := BaseVestingAccount{
		OriginalVesting:  sdk.NewCoins(),
		DelegatedFree:    sdk.NewCoins(),
		DelegatedVesting: sdk.NewCoins(),
		AddressCodec:     d.AddressCodec,
		EndTime:          0,
	}

	continuousVestingAccount := ContinuousVestingAccount{
		BaseVestingAccount: &baseVestingAccount,
		StartTime:          0,
	}

	return &continuousVestingAccount, nil
}

type ContinuousVestingAccount struct {
	*BaseVestingAccount
	StartTime int64
}

// --------------- Init -----------------

func (cva ContinuousVestingAccount) Init(ctx context.Context, msg *vestingtypes.MsgInitVestingAccount) (*vestingtypes.MsgInitVestingAccountResponse, error) {
	to := accountstd.Whoami(ctx)
	if to == nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("Cannot find account address from context")
	}

	toAddress, err := cva.AddressCodec.BytesToString(to)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid 'to' address: %s", err)
	}

	if err := validateAmount(msg.Amount); err != nil {
		return nil, err
	}

	if msg.StartTime < 1 {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid start time of %d, length must be greater than 0", msg.StartTime)
	}

	if msg.EndTime <= 0 {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("invalid end time")
	}

	if msg.EndTime <= msg.StartTime {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("invalid start and end time (must be start < end)")
	}

	originalContext := accountstd.OriginalContext(ctx)

	start := msg.StartTime
	if msg.StartTime == 0 {
		sdkctx := sdk.UnwrapSDKContext(originalContext)
		start = sdkctx.HeaderInfo().Time.Unix()
	}

	cva.OriginalVesting = msg.Amount.Sort()
	cva.DelegatedFree = sdk.NewCoins()
	cva.DelegatedVesting = sdk.NewCoins()
	cva.EndTime = msg.EndTime
	cva.StartTime = start

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
	err = cva.Validate()
	if err != nil {
		return nil, err
	}

	return &vestingtypes.MsgInitVestingAccountResponse{}, nil
}

// --------------- execute -----------------
// LockedCoins returns the set of coins that are not spendable (i.e. locked),
// defined as the vesting coins that are not delegated.
func (cva ContinuousVestingAccount) LockedCoins(blockTime time.Time) sdk.Coins {
	return cva.BaseVestingAccount.LockedCoinsFromVesting(cva.GetVestingCoins(blockTime))
}

func (cva *ContinuousVestingAccount) ExecuteMessages(ctx context.Context, msg *vestingtypes.MsgExecuteMessages) (
	*vestingtypes.MsgExecuteMessagesResponse, error,
) {
	return cva.BaseVestingAccount.ExecuteMessages(ctx, msg, cva.GetVestingCoins)
}

// --------------- Query -----------------

// GetVestedCoins returns the total number of vested coins. If no coins are vested,
// nil is returned.
func (cva ContinuousVestingAccount) GetVestedCoins(blockTime time.Time) sdk.Coins {
	var vestedCoins sdk.Coins

	// We must handle the case where the start time for a vesting account has
	// been set into the future or when the start of the chain is not exactly
	// known.
	if blockTime.Unix() <= cva.StartTime {
		return vestedCoins
	} else if blockTime.Unix() >= cva.EndTime {
		return cva.OriginalVesting
	}

	// calculate the vesting scalar
	x := blockTime.Unix() - cva.StartTime
	y := cva.EndTime - cva.StartTime
	s := math.LegacyNewDec(x).Quo(math.LegacyNewDec(y))

	for _, ovc := range cva.OriginalVesting {
		vestedAmt := math.LegacyNewDecFromInt(ovc.Amount).Mul(s).RoundInt()
		vestedCoins = append(vestedCoins, sdk.NewCoin(ovc.Denom, vestedAmt))
	}

	return vestedCoins
}

// GetVestingCoins returns the total number of vesting coins. If no coins are
// vesting, nil is returned.
func (cva ContinuousVestingAccount) GetVestingCoins(blockTime time.Time) sdk.Coins {
	return cva.OriginalVesting.Sub(cva.GetVestedCoins(blockTime)...)
}

func (cva ContinuousVestingAccount) QueryVestedCoins(ctx context.Context, msg *vestingtypes.QueryVestedCoinsRequest) (
	*vestingtypes.QueryVestedCoinsResponse, error,
) {
	originalContext := accountstd.OriginalContext(ctx)
	sdkctx := sdk.UnwrapSDKContext(originalContext)
	vestedCoins := cva.GetVestedCoins(sdkctx.HeaderInfo().Time)

	return &vestingtypes.QueryVestedCoinsResponse{
		VestedVesting: vestedCoins,
	}, nil
}

func (cva ContinuousVestingAccount) QueryVestingCoins(ctx context.Context, msg *vestingtypes.QueryVestingCoinsRequest) (
	*vestingtypes.QueryVestingCoinsResponse, error,
) {
	originalContext := accountstd.OriginalContext(ctx)
	sdkctx := sdk.UnwrapSDKContext(originalContext)
	vestingCoins := cva.GetVestingCoins(sdkctx.BlockHeader().Time)

	return &vestingtypes.QueryVestingCoinsResponse{
		VestingCoins: vestingCoins,
	}, nil
}

func (cva ContinuousVestingAccount) QueryStartTime(ctx context.Context, msg *vestingtypes.QueryStartTimeRequest) (
	*vestingtypes.QueryStartTimeResponse, error,
) {
	return &vestingtypes.QueryStartTimeResponse{
		StartTime: cva.StartTime,
	}, nil
}

// Validate checks for errors on the account fields
func (cva ContinuousVestingAccount) Validate() error {
	if cva.StartTime >= cva.EndTime {
		return errors.New("vesting start-time cannot be before end-time")
	}

	return cva.BaseVestingAccount.Validate()
}

// Implement smart account interface
func (cva ContinuousVestingAccount) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, cva.Init)
}

func (cva ContinuousVestingAccount) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, cva.ExecuteMessages)
}

func (cva ContinuousVestingAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, cva.QueryStartTime)
	accountstd.RegisterQueryHandler(builder, cva.QueryVestedCoins)
	accountstd.RegisterQueryHandler(builder, cva.QueryVestingCoins)
	cva.BaseVestingAccount.RegisterQueryHandlers(builder)
}
