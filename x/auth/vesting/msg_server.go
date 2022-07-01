package vesting

import (
	"context"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/armon/go-metrics"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

type msgServer struct {
	keeper.AccountKeeper
	types.BankKeeper
	types.DistrKeeper
}

// NewMsgServerImpl returns an implementation of the vesting MsgServer interface,
// wrapping the corresponding AccountKeeper and BankKeeper.
func NewMsgServerImpl(k keeper.AccountKeeper, bk types.BankKeeper, dk types.DistrKeeper) types.MsgServer {
	return &msgServer{AccountKeeper: k, BankKeeper: bk, DistrKeeper: dk}
}

var _ types.MsgServer = msgServer{}

func (s msgServer) CreateVestingAccount(goCtx context.Context, msg *types.MsgCreateVestingAccount) (*types.MsgCreateVestingAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ak := s.AccountKeeper
	bk := s.BankKeeper

	if err := bk.IsSendEnabledCoins(ctx, msg.Amount...); err != nil {
		return nil, err
	}

	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, err
	}
	to, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, err
	}

	if bk.BlockedAddr(to) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", msg.ToAddress)
	}

	if acc := ak.GetAccount(ctx, to); acc != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s already exists", msg.ToAddress)
	}

	baseAccount := ak.NewAccountWithAddress(ctx, to)
	if _, ok := baseAccount.(*authtypes.BaseAccount); !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid account type; expected: BaseAccount, got: %T", baseAccount)
	}

	baseVestingAccount := types.NewBaseVestingAccount(baseAccount.(*authtypes.BaseAccount), msg.Amount.Sort(), msg.EndTime)

	var acc authtypes.AccountI

	if msg.Delayed {
		acc = types.NewDelayedVestingAccountRaw(baseVestingAccount)
	} else {
		acc = types.NewContinuousVestingAccountRaw(baseVestingAccount, ctx.BlockTime().Unix())
	}

	ak.SetAccount(ctx, acc)

	defer func() {
		telemetry.IncrCounter(1, "new", "account")

		for _, a := range msg.Amount {
			if a.Amount.IsInt64() {
				telemetry.SetGaugeWithLabels(
					[]string{"tx", "msg", "create_vesting_account"},
					float32(a.Amount.Int64()),
					[]metrics.Label{telemetry.NewLabel("denom", a.Denom)},
				)
			}
		}
	}()

	err = bk.SendCoins(ctx, from, to, msg.Amount)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	return &types.MsgCreateVestingAccountResponse{}, nil
}

func (s msgServer) CreatePeriodicVestingAccount(goCtx context.Context, msg *types.MsgCreatePeriodicVestingAccount) (*types.MsgCreatePeriodicVestingAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ak := s.AccountKeeper
	bk := s.BankKeeper

	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, err
	}
	to, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, err
	}

	if acc := ak.GetAccount(ctx, to); acc != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s already exists", msg.ToAddress)
	}

	var totalCoins sdk.Coins

	for _, period := range msg.VestingPeriods {
		totalCoins = totalCoins.Add(period.Amount...)
	}

	baseAccount := ak.NewAccountWithAddress(ctx, to)

	acc := types.NewPeriodicVestingAccount(baseAccount.(*authtypes.BaseAccount), totalCoins.Sort(), msg.StartTime, msg.VestingPeriods)

	ak.SetAccount(ctx, acc)

	defer func() {
		telemetry.IncrCounter(1, "new", "account")

		for _, a := range totalCoins {
			if a.Amount.IsInt64() {
				telemetry.SetGaugeWithLabels(
					[]string{"tx", "msg", "create_periodic_vesting_account"},
					float32(a.Amount.Int64()),
					[]metrics.Label{telemetry.NewLabel("denom", a.Denom)},
				)
			}
		}
	}()

	err = bk.SendCoins(ctx, from, to, totalCoins)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)
	return &types.MsgCreatePeriodicVestingAccountResponse{}, nil

}

func (s msgServer) DonateAllVestingTokens(goCtx context.Context, msg *types.MsgDonateAllVestingTokens) (*types.MsgDonateAllVestingTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ak := s.AccountKeeper
	dk := s.DistrKeeper

	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, err
	}

	acc := ak.GetAccount(ctx, from)
	if acc == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s not exists", msg.FromAddress)
	}

	vestingAcc, ok := acc.(exported.VestingAccount)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s is not vesting account", msg.FromAddress)
	}

	if !vestingAcc.GetDelegatedVesting().IsZero() {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s has delegated vesting tokens", msg.FromAddress)
	}

	vestingCoins := vestingAcc.GetVestingCoins(ctx.BlockTime())
	if vestingCoins.IsZero() {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s has no vesting tokens", msg.FromAddress)
	}

	// Change the account as normal account
	ak.SetAccount(ctx,
		authtypes.NewBaseAccount(
			acc.GetAddress(),
			acc.GetPubKey(),
			acc.GetAccountNumber(),
			acc.GetSequence(),
		),
	)

	if err := dk.FundCommunityPool(ctx, vestingCoins, from); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	return &types.MsgDonateAllVestingTokensResponse{}, nil
}
