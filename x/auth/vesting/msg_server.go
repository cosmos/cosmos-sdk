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

	if err := bk.SendEnabledCoins(ctx, msg.Amount...); err != nil {
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

func min64(i, j int64) int64 {
	if i < j {
		return i
	}
	return j
}

// mergePeriods returns the merge of two vesting period schedules.
// The merge is defined as the union of the vesting events, with simultaneous
// events combined into a single event.
// Returns new start time, new end time, and merged vesting events, relative to
// the new start time.
func mergePeriods(startP, startQ int64, p, q []types.Period) (int64, int64, []types.Period) {
	timeP := startP // time of last merged p event, next p event is relative to this time
	timeQ := startQ // time of last merged q event, next q event is relative to this time
	iP := 0         // p indexes before this have been merged
	iQ := 0         // q indexes before this have been merged
	lenP := len(p)
	lenQ := len(q)
	startTime := min64(startP, startQ) // we pick the earlier time
	time := startTime                  // time of last merged event, or the start time
	merged := []types.Period{}

	// emit adds a merged period and updates the last event time
	emit := func(nextTime int64, amount sdk.Coins) {
		period := types.Period{
			Length: nextTime - time,
			Amount: amount,
		}
		merged = append(merged, period)
		time = nextTime
	}

	// consumeP emits the next period from p, updating indexes
	consumeP := func(nextP int64) {
		emit(nextP, p[iP].Amount)
		timeP = nextP
		iP++
	}

	// consumeQ emits the next period from q, updating indexes
	consumeQ := func(nextQ int64) {
		emit(nextQ, q[iQ].Amount)
		timeQ = nextQ
		iQ++
	}

	// consumeBoth emits a merge of the next periods from p and q, updating indexes
	consumeBoth := func(nextTime int64) {
		emit(nextTime, p[iP].Amount.Add(q[iQ].Amount...))
		timeP = nextTime
		timeQ = nextTime
		iP++
		iQ++
	}

	for iP < lenP && iQ < lenQ {
		nextP := timeP + p[iP].Length // next p event in absolute time
		nextQ := timeQ + q[iQ].Length // next q event in absolute time
		if nextP < nextQ {
			consumeP(nextP)
		} else if nextP > nextQ {
			consumeQ(nextQ)
		} else {
			consumeBoth(nextP)
		}
	}
	for iP < lenP {
		// Ragged end - consume remaining p
		nextP := timeP + p[iP].Length
		consumeP(nextP)
	}
	for iQ < lenQ {
		// Ragged end - consume remaining q
		nextQ := timeQ + q[iQ].Length
		consumeQ(nextQ)
	}
	return startTime, time, merged
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

	if bk.BlockedAddr(to) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", msg.ToAddress)
	}

	var totalCoins sdk.Coins
	for _, period := range msg.VestingPeriods {
		totalCoins = totalCoins.Add(period.Amount...)
	}
	totalCoins = totalCoins.Sort()

	acc := ak.GetAccount(ctx, to)

	if acc != nil {
		pva, ok := acc.(*types.PeriodicVestingAccount)
		if !msg.Merge {
			if ok {
				return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s already exists; consider using --merge", msg.ToAddress)
			}
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s already exists", msg.ToAddress)
		}
		if !ok {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrNotSupported, "account %s must be a periodic vestic account", msg.ToAddress)
		}
		newStart, newEnd, newPeriods := mergePeriods(pva.StartTime, msg.GetStartTime(),
			pva.GetVestingPeriods(), msg.GetVestingPeriods())
		pva.StartTime = newStart
		pva.EndTime = newEnd
		pva.VestingPeriods = newPeriods
		pva.OriginalVesting = pva.OriginalVesting.Add(totalCoins...)
	} else {
		baseAccount := ak.NewAccountWithAddress(ctx, to)
		acc = types.NewPeriodicVestingAccount(baseAccount.(*authtypes.BaseAccount), totalCoins, msg.StartTime, msg.VestingPeriods)
	}

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
