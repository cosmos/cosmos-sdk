package vesting

import (
	"context"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/armon/go-metrics"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

type msgServer struct {
	keeper.AccountKeeper
	types.BankKeeper
}

// NewMsgServerImpl returns an implementation of the vesting MsgServer interface,
// wrapping the corresponding AccountKeeper and BankKeeper.
func NewMsgServerImpl(k keeper.AccountKeeper, bk types.BankKeeper) types.MsgServer {
	return &msgServer{AccountKeeper: k, BankKeeper: bk}
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

// CreateClawbackVestingAccount creates a new ClawbackVestingAccount, or merges a grant into an existing one.
func (s msgServer) CreateClawbackVestingAccount(goCtx context.Context, msg *types.MsgCreateClawbackVestingAccount) (*types.MsgCreateClawbackVestingAccountResponse, error) {
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

	if bk.BlockedAddr(to) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", msg.ToAddress)
	}

	vestingCoins := sdk.NewCoins()
	for _, period := range msg.VestingPeriods {
		vestingCoins = vestingCoins.Add(period.Amount...)
	}

	lockupCoins := sdk.NewCoins()
	for _, period := range msg.LockupPeriods {
		lockupCoins = lockupCoins.Add(period.Amount...)
	}

	// if lockup absent, default to an instant unlock schedule
	lockupPeriods := msg.LockupPeriods
	vestingPeriods := msg.VestingPeriods
	if !vestingCoins.IsZero() && len(msg.LockupPeriods) == 0 {
		lockupPeriods = []types.Period{
			{Length: 0, Amount: vestingCoins},
		}
		lockupCoins = vestingCoins
	}

	if !lockupCoins.IsZero() && len(msg.VestingPeriods) == 0 {
		// If vesting absent, default to an instant vesting schedule
		vestingPeriods = []types.Period{
			{Length: 0, Amount: lockupCoins},
		}
		vestingCoins = lockupCoins
	}

	if !types.CoinEq(lockupCoins, vestingCoins) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "lockup and vesting amounts must be equal")
	}

	var (
		madeNewAcc bool
		vestingAcc *types.ClawbackVestingAccount
	)

	acc := ak.GetAccount(ctx, to)

	// a grant can be added only if toAddress exists, msg.Merge && isClawback && to.FunderAddress == msg.FromAddress
	if acc != nil {
		var isClawback bool
		vestingAcc, isClawback = acc.(*types.ClawbackVestingAccount)
		switch {
		case !isClawback:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrNotSupported, "account %s must be a clawback vesting account", msg.ToAddress)
		case !msg.Merge && isClawback:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s already exists; consider setting  'merge' to 'true'", msg.ToAddress)
		case msg.FromAddress != vestingAcc.FunderAddress:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s can only accept grants from account %s", msg.ToAddress, vestingAcc.FunderAddress)
		}
		grantAction := types.NewClawbackGrantAction(msg.FromAddress, msg.StartTime, msg.GetLockupPeriods(), msg.GetVestingPeriods(), vestingCoins)
		err := vestingAcc.AddGrant(ctx, grantAction)
		if err != nil {
			return nil, err
		}
		ak.SetAccount(ctx, vestingAcc)
	} else {
		baseAcc := authtypes.NewBaseAccountWithAddress(to)
		vestingAcc = types.NewClawbackVestingAccount(
			baseAcc,
			from,
			vestingCoins,
			msg.StartTime,
			lockupPeriods,
			vestingPeriods,
		)
		acc := ak.NewAccount(ctx, vestingAcc)
		madeNewAcc = true
		ak.SetAccount(ctx, acc)
	}

	if madeNewAcc {
		defer func() {
			telemetry.IncrCounter(1, "new", "account")

			for _, a := range vestingCoins {
				if a.Amount.IsInt64() {
					telemetry.SetGaugeWithLabels(
						[]string{"tx", "msg", "create_clawback_vesting_account"},
						float32(a.Amount.Int64()),
						[]metrics.Label{telemetry.NewLabel("denom", a.Denom)},
					)
				}
			}
		}()
	}

	err = bk.SendCoins(ctx, from, to, vestingCoins)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.FromAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, vestingCoins.String()),
		),
	)

	return &types.MsgCreateClawbackVestingAccountResponse{}, nil
}

// Clawback removes the unvested amount from a ClawbackVestingAccount.
// The destination defaults to the funder address, but can be overridden.
func (s msgServer) Clawback(goCtx context.Context, msg *types.MsgClawback) (*types.MsgClawbackResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	accountKeeper := s.AccountKeeper
	bankKeeper := s.BankKeeper

	funder, err := sdk.AccAddressFromBech32(msg.GetFunderAddress())
	if err != nil {
		return nil, err
	}

	addr, err := sdk.AccAddressFromBech32(msg.GetAddress())
	if err != nil {
		return nil, err
	}

	dest := funder
	if msg.GetDestAddress() != "" {
		dest, err = sdk.AccAddressFromBech32(msg.GetDestAddress())
		if err != nil {
			return nil, err
		}
	}

	if bankKeeper.BlockedAddr(dest) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized,
			"%s is not allowed to receive funds", msg.DestAddress,
		)
	}

	// Check if account exists
	account := accountKeeper.GetAccount(ctx, addr)
	if account == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "account %s does not exist", msg.Address)
	}

	// Check if account has a clawback account
	vestingAccount, ok := account.(*types.ClawbackVestingAccount)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account not subject to clawback: %s", msg.Address)
	}

	// Check if account funder is same as in msg
	if vestingAccount.FunderAddress != msg.FunderAddress {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "clawback can only be requested by original funder %s", vestingAccount.FunderAddress)
	}

	clawbackAction := types.NewClawbackAction(funder, dest, accountKeeper, bankKeeper)

	// Perform clawback transfer,
	// this updates state for both the vesting account and the destination account.
	err = vestingAccount.Clawback(ctx, clawbackAction)
	if err != nil {
		return nil, err
	}

	return &types.MsgClawbackResponse{}, nil
}
