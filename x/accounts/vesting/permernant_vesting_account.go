package vesting

import (
	"context"
	"errors"
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
	_ accountstd.Interface = (*PermanentLockedAccount)(nil)
)

// Base Vesting Account

// NewPermanentLockedAccount creates a new PermanentLockedAccount object. It is the
// callers responsibility to ensure the base account has sufficient funds with
// regards to the original vesting amount.
func NewPermanentLockedAccount(d accountstd.Dependencies) (*PermanentLockedAccount, error) {
	baseVestingAccount := BaseVestingAccount{
		OriginalVesting:  sdk.NewCoins(),
		DelegatedFree:    sdk.NewCoins(),
		DelegatedVesting: sdk.NewCoins(),
		AddressCodec:     d.AddressCodec,
		EndTime:          0,
	}

	return &PermanentLockedAccount{&baseVestingAccount}, nil
}

type PermanentLockedAccount struct {
	*BaseVestingAccount
}

// --------------- Init -----------------

func (cva PermanentLockedAccount) Init(ctx context.Context, msg *vestingtypes.MsgInitVestingAccount) (*vestingtypes.MsgInitVestingAccountResponse, error) {
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

	cva.OriginalVesting = msg.Amount.Sort()
	cva.DelegatedFree = sdk.NewCoins()
	cva.DelegatedVesting = sdk.NewCoins()

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
func (plva PermanentLockedAccount) LockedCoins(_ time.Time) sdk.Coins {
	return plva.BaseVestingAccount.LockedCoinsFromVesting(plva.OriginalVesting)
}

// TrackDelegation tracks a desired delegation amount by setting the appropriate
// values for the amount of delegated vesting, delegated free, and reducing the
// overall amount of base coins.
func (plva *PermanentLockedAccount) TrackDelegation(blockTime time.Time, balance, amount sdk.Coins) {
	plva.BaseVestingAccount.TrackDelegation(balance, plva.OriginalVesting, amount)
}

// --------------- Query -----------------
// GetVestedCoins returns the total amount of vested coins for a permanent locked vesting
// account. All coins are only vested once the schedule has elapsed.
func (plva PermanentLockedAccount) GetVestedCoins(_ time.Time) sdk.Coins {
	return nil
}

// GetVestingCoins returns the total number of vesting coins for a permanent locked
// vesting account.
func (plva PermanentLockedAccount) GetVestingCoins(_ time.Time) sdk.Coins {
	return plva.OriginalVesting
}

func (plva PermanentLockedAccount) QueryVestedCoins(ctx context.Context, msg *vestingtypes.QueryVestedCoinsRequest) (
	*vestingtypes.QueryVestedCoinsResponse, error,
) {
	originalContext := accountstd.OriginalContext(ctx)
	sdkctx := sdk.UnwrapSDKContext(originalContext)
	vestedCoins := plva.GetVestedCoins(sdkctx.HeaderInfo().Time)

	return &vestingtypes.QueryVestedCoinsResponse{
		VestedVesting: vestedCoins,
	}, nil
}

func (plva PermanentLockedAccount) QueryVestingCoins(ctx context.Context, msg *vestingtypes.QueryVestingCoinsRequest) (
	*vestingtypes.QueryVestingCoinsResponse, error,
) {
	originalContext := accountstd.OriginalContext(ctx)
	sdkctx := sdk.UnwrapSDKContext(originalContext)
	vestingCoins := plva.GetVestingCoins(sdkctx.BlockHeader().Time)

	return &vestingtypes.QueryVestingCoinsResponse{
		VestingCoins: vestingCoins,
	}, nil
}

// Validate checks for errors on the account fields
func (plva PermanentLockedAccount) Validate() error {
	if plva.EndTime > 0 {
		return errors.New("permanently vested accounts cannot have an end-time")
	}

	return plva.BaseVestingAccount.Validate()
}

// Implement smart account interface
func (plva PermanentLockedAccount) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, plva.Init)
}

func (plva PermanentLockedAccount) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
}

func (plva PermanentLockedAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, plva.QueryOriginalVesting)
	accountstd.RegisterQueryHandler(builder, plva.QueryDelegatedFree)
	accountstd.RegisterQueryHandler(builder, plva.QueryDelegatedVesting)
	accountstd.RegisterQueryHandler(builder, plva.QueryEndTime)
	accountstd.RegisterQueryHandler(builder, plva.QueryVestedCoins)
	accountstd.RegisterQueryHandler(builder, plva.QueryVestingCoins)
}
