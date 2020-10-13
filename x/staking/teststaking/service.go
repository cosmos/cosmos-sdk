package teststaking

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Service is a structure which wraps the staking handler
// and provides methods useful in tests
type Service struct {
	t *testing.T
	h sdk.Handler
	k keeper.Keeper

	Ctx        sdk.Context
	Commission stakingtypes.CommissionRates
	// Coin Denomination
	Denom string
}

// NewService creates staking Handler wrapper for tests
func NewService(t *testing.T, ctx sdk.Context, k keeper.Keeper) *Service {
	return &Service{t, staking.NewHandler(k), k, ctx, ZeroCommission(), sdk.DefaultBondDenom}
}

// CreateValidator calls handler to create a new staking validator
func (sh *Service) CreateValidator(addr sdk.ValAddress, pk crypto.PubKey, stakeAmount int64, ok bool) {
	coin := sdk.NewCoin(sh.Denom, sdk.NewInt(stakeAmount))
	sh.createValidator(addr, pk, coin, ok)
}

// CreateValidatorWithValPower calls handler to create a new staking validator with zero
// commission
func (sh *Service) CreateValidatorWithValPower(addr sdk.ValAddress, pk crypto.PubKey, valPower int64, ok bool) sdk.Int {
	amount := sdk.TokensFromConsensusPower(valPower)
	coin := sdk.NewCoin(sh.Denom, amount)
	sh.createValidator(addr, pk, coin, ok)
	return amount
}

// CreateValidatorMsg returns a message used to create validator in this service.
func (sh *Service) CreateValidatorMsg(addr sdk.ValAddress, pk crypto.PubKey, stakeAmount int64) *stakingtypes.MsgCreateValidator {
	coin := sdk.NewCoin(sh.Denom, sdk.NewInt(stakeAmount))
	msg, err := stakingtypes.NewMsgCreateValidator(addr, pk, coin, stakingtypes.Description{}, sh.Commission, sdk.OneInt())
	require.NoError(sh.t, err)
	return msg
}

func (sh *Service) createValidator(addr sdk.ValAddress, pk crypto.PubKey, coin sdk.Coin, ok bool) {
	msg, err := stakingtypes.NewMsgCreateValidator(addr, pk, coin, stakingtypes.Description{}, sh.Commission, sdk.OneInt())
	require.NoError(sh.t, err)
	sh.Handle(msg, ok)
}

// Delegate calls handler to delegate stake for a validator
func (sh *Service) Delegate(delegator sdk.AccAddress, val sdk.ValAddress, amount int64) {
	coin := sdk.NewCoin(sh.Denom, sdk.NewInt(amount))
	msg := stakingtypes.NewMsgDelegate(delegator, val, coin)
	sh.Handle(msg, true)
}

// DelegateWithPower calls handler to delegate stake for a validator
func (sh *Service) DelegateWithPower(delegator sdk.AccAddress, val sdk.ValAddress, power int64) {
	coin := sdk.NewCoin(sh.Denom, sdk.TokensFromConsensusPower(power))
	msg := stakingtypes.NewMsgDelegate(delegator, val, coin)
	sh.Handle(msg, true)
}

// Undelegate calls handler to unbound some stake from a validator.
func (sh *Service) Undelegate(delegator sdk.AccAddress, val sdk.ValAddress, amount sdk.Int, ok bool) *sdk.Result {
	unbondAmt := sdk.NewCoin(sh.Denom, amount)
	msg := stakingtypes.NewMsgUndelegate(delegator, val, unbondAmt)
	return sh.Handle(msg, ok)
}

// Handle calls staking handler on a given message
func (sh *Service) Handle(msg sdk.Msg, ok bool) *sdk.Result {
	res, err := sh.h(sh.Ctx, msg)
	if ok {
		require.NoError(sh.t, err)
		require.NotNil(sh.t, res)
	} else {
		require.Error(sh.t, err)
		require.Nil(sh.t, res)
	}
	return res
}

// CheckValidator asserts that a validor exists and has a given status (if status!="")
// and if has a right jailed flag.
func (sh *Service) CheckValidator(addr sdk.ValAddress, status stakingtypes.BondStatus, jailed bool) stakingtypes.Validator {
	v, ok := sh.k.GetValidator(sh.Ctx, addr)
	require.True(sh.t, ok)
	require.Equal(sh.t, jailed, v.Jailed, "wrong Jalied status")
	if status >= 0 {
		require.Equal(sh.t, status, v.Status)
	}
	return v
}

// CheckDelegator asserts that a delegator exists
func (sh *Service) CheckDelegator(delegator sdk.AccAddress, val sdk.ValAddress) {
	d, ok := sh.k.GetDelegation(sh.Ctx, delegator, val)
	require.True(sh.t, ok)
	require.NotNil(sh.t, d)
}

// TurnBlock calls EndBlocker and updates the block time
func (sh *Service) TurnBlock(newTime time.Time) sdk.Context {
	sh.Ctx = sh.Ctx.WithBlockTime(newTime)
	staking.EndBlocker(sh.Ctx, sh.k)
	return sh.Ctx
}

// TurnBlockTimeDiff calls EndBlocker and updates the block time by adding the
// duration to the current block time
func (sh *Service) TurnBlockTimeDiff(diff time.Duration) sdk.Context {
	sh.Ctx = sh.Ctx.WithBlockTime(sh.Ctx.BlockHeader().Time.Add(diff))
	staking.EndBlocker(sh.Ctx, sh.k)
	return sh.Ctx
}

// ZeroCommission constructs a commission rates with all zeros.
func ZeroCommission() stakingtypes.CommissionRates {
	return stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
}
