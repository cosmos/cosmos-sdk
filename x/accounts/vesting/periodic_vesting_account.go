package vesting

import (
	"context"
	"errors"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/x/accounts/accountstd"
	vestingtypes "cosmossdk.io/x/accounts/vesting/types/v1"
	banktypes "cosmossdk.io/x/bank/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Compile-time type assertions
var (
	_ accountstd.Interface = (*PeriodicVestingAccount)(nil)
)

// Periodic Vesting Account

// NewPeriodicVestingAccount creates a new PeriodicVestingAccount object.
func NewPeriodicVestingAccount(d accountstd.Dependencies) (*PeriodicVestingAccount, error) {
	baseVestingAccount := BaseVestingAccount{
		OriginalVesting:  sdk.NewCoins(),
		DelegatedFree:    sdk.NewCoins(),
		DelegatedVesting: sdk.NewCoins(),
		AddressCodec:     d.AddressCodec,
		EndTime:          0,
	}

	periodicsVestingAccount := PeriodicVestingAccount{
		BaseVestingAccount: &baseVestingAccount,
		StartTime:          0,
		VestingPeriods:     []vestingtypes.Period{},
	}

	return &periodicsVestingAccount, nil
}

type PeriodicVestingAccount struct {
	*BaseVestingAccount
	StartTime      int64
	VestingPeriods []vestingtypes.Period
}

// --------------- Init -----------------

func (pva PeriodicVestingAccount) Init(ctx context.Context, msg *vestingtypes.MsgInitPeriodicVestingAccount) (*vestingtypes.MsgInitPeriodicVestingAccountResponse, error) {
	to := accountstd.Whoami(ctx)
	if to == nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("Cannot find account address from context")
	}

	toAddress, err := pva.AddressCodec.BytesToString(to)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid 'to' address: %s", err)
	}

	if msg.StartTime < 1 {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid start time of %d, length must be greater than 0", msg.StartTime)
	}

	var totalCoins sdk.Coins
	endTime := msg.StartTime
	for i, period := range msg.VestingPeriods {
		if period.Length < 1 {
			return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid period length of %d in period %d, length must be greater than 0", period.Length, i)
		}

		if err := validateAmount(period.Amount); err != nil {
			return nil, err
		}

		totalCoins = totalCoins.Add(period.Amount...)
		// Calculate end time
		endTime += period.Length
	}

	pva.OriginalVesting = totalCoins.Sort()
	pva.DelegatedFree = sdk.NewCoins()
	pva.DelegatedVesting = sdk.NewCoins()
	pva.StartTime = msg.StartTime
	pva.EndTime = endTime

	// Send token to new vesting account
	sendMsg := banktypes.NewMsgSend(msg.FromAddress, toAddress, totalCoins)
	anyMsg, err := codectypes.NewAnyWithValue(sendMsg)
	if err != nil {
		return nil, err
	}

	if _, err = accountstd.ExecModuleAnys(ctx, []*codectypes.Any{anyMsg}); err != nil {
		return nil, err
	}

	// Validate the newly init account
	err = pva.Validate()
	if err != nil {
		return nil, err
	}

	return &vestingtypes.MsgInitPeriodicVestingAccountResponse{}, nil
}

// --------------- execute -----------------

// LockedCoins returns the set of coins that are not spendable (i.e. locked),
// defined as the vesting coins that are not delegated.
func (pva PeriodicVestingAccount) LockedCoins(blockTime time.Time) sdk.Coins {
	return pva.BaseVestingAccount.LockedCoinsFromVesting(pva.GetVestingCoins(blockTime))
}

// TrackDelegation tracks a desired delegation amount by setting the appropriate
// values for the amount of delegated vesting, delegated free, and reducing the
// overall amount of base coins.
func (pva *PeriodicVestingAccount) TrackDelegation(blockTime time.Time, balance, amount sdk.Coins) {
	pva.BaseVestingAccount.TrackDelegation(balance, pva.GetVestingCoins(blockTime), amount)
}

// Validate checks for errors on the account fields
func (pva PeriodicVestingAccount) Validate() error {
	if pva.StartTime >= pva.EndTime {
		return errors.New("vesting start-time cannot be before end-time")
	}

	return pva.BaseVestingAccount.Validate()
}

// ----------------- Query --------------------

// GetVestedCoins returns the total number of vested coins. If no coins are vested,
// nil is returned.
func (pva PeriodicVestingAccount) GetVestedCoins(blockTime time.Time) sdk.Coins {
	var vestedCoins sdk.Coins

	// We must handle the case where the start time for a vesting account has
	// been set into the future or when the start of the chain is not exactly
	// known.
	if blockTime.Unix() <= pva.StartTime {
		return vestedCoins
	} else if blockTime.Unix() >= pva.EndTime {
		return pva.OriginalVesting
	}

	// track the start time of the next period
	currentPeriodStartTime := pva.StartTime

	// for each period, if the period is over, add those coins as vested and check the next period.
	for _, period := range pva.VestingPeriods {
		x := blockTime.Unix() - currentPeriodStartTime
		if x < period.Length {
			break
		}

		vestedCoins = vestedCoins.Add(period.Amount...)

		// update the start time of the next period
		currentPeriodStartTime += period.Length
	}

	return vestedCoins
}

// GetVestingCoins returns the total number of vesting coins. If no coins are
// vesting, nil is returned.
func (pva PeriodicVestingAccount) GetVestingCoins(blockTime time.Time) sdk.Coins {
	return pva.OriginalVesting.Sub(pva.GetVestedCoins(blockTime)...)
}

func (pva PeriodicVestingAccount) QueryVestedCoins(ctx context.Context, msg *vestingtypes.QueryVestedCoinsRequest) (
	*vestingtypes.QueryVestedCoinsResponse, error,
) {
	originalContext := accountstd.OriginalContext(ctx)
	sdkctx := sdk.UnwrapSDKContext(originalContext)
	vestedCoins := pva.GetVestedCoins(sdkctx.HeaderInfo().Time)

	return &vestingtypes.QueryVestedCoinsResponse{
		VestedVesting: vestedCoins,
	}, nil
}

func (pva PeriodicVestingAccount) QueryVestingCoins(ctx context.Context, msg *vestingtypes.QueryVestingCoinsRequest) (
	*vestingtypes.QueryVestingCoinsResponse, error,
) {
	originalContext := accountstd.OriginalContext(ctx)
	sdkctx := sdk.UnwrapSDKContext(originalContext)
	vestingCoins := pva.GetVestingCoins(sdkctx.BlockHeader().Time)

	return &vestingtypes.QueryVestingCoinsResponse{
		VestingCoins: vestingCoins,
	}, nil
}

func (pva PeriodicVestingAccount) QueryStartTime(ctx context.Context, msg *vestingtypes.QueryStartTimeRequest) (
	*vestingtypes.QueryStartTimeResponse, error,
) {
	return &vestingtypes.QueryStartTimeResponse{
		StartTime: pva.StartTime,
	}, nil
}

// Implement smart account interface
func (pva PeriodicVestingAccount) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, pva.Init)
}

func (pva PeriodicVestingAccount) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
}

func (pva PeriodicVestingAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, pva.QueryStartTime)
	accountstd.RegisterQueryHandler(builder, pva.QueryVestedCoins)
	accountstd.RegisterQueryHandler(builder, pva.QueryVestingCoins)
	pva.BaseVestingAccount.RegisterQueryHandlers(builder)
}
