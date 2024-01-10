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

// Permernant Vesting Account

// NewPermanentLockedAccount creates a new PermanentLockedAccount object.
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

func (plva *PermanentLockedAccount) ExecuteMessages(ctx context.Context, msg *vestingtypes.MsgExecuteMessages) (
	*vestingtypes.MsgExecuteMessagesResponse, error,
) {
	return plva.BaseVestingAccount.ExecuteMessages(ctx, msg, func(_ time.Time) sdk.Coins {
		return plva.OriginalVesting
	})
}

// --------------- Query -----------------

func (plva PermanentLockedAccount) QueryVestedCoins(ctx context.Context, msg *vestingtypes.QueryVestedCoinsRequest) (
	*vestingtypes.QueryVestedCoinsResponse, error,
) {
	return &vestingtypes.QueryVestedCoinsResponse{
		VestedVesting: nil,
	}, nil
}

func (plva PermanentLockedAccount) QueryVestingCoins(ctx context.Context, msg *vestingtypes.QueryVestingCoinsRequest) (
	*vestingtypes.QueryVestingCoinsResponse, error,
) {
	return &vestingtypes.QueryVestingCoinsResponse{
		VestingCoins: plva.OriginalVesting,
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
	accountstd.RegisterExecuteHandler(builder, plva.ExecuteMessages)
}

func (plva PermanentLockedAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, plva.QueryVestedCoins)
	accountstd.RegisterQueryHandler(builder, plva.QueryVestingCoins)
	plva.BaseVestingAccount.RegisterQueryHandlers(builder)
}
