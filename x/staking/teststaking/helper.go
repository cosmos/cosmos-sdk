package teststaking

import (
	"testing"
	"time"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

// Helper is a structure which wraps the staking message server
// and provides methods useful in tests
type Helper struct {
	t       *testing.T
	MsgSrvr stakingtypes.MsgServer
	k       keeper.Keeper

	Ctx        sdk.Context
	Commission stakingtypes.CommissionRates
	// Coin Denomination
	Denom string
}

// NewHelper creates a new instance of Helper.
func NewHelper(t *testing.T, ctx sdk.Context, k keeper.Keeper) *Helper {
	return &Helper{t, keeper.NewMsgServerImpl(k), k, ctx, ZeroCommission(), sdk.DefaultBondDenom}
}

// CreateValidator calls staking module `MsgServer/CreateValidator` to create a new validator
func (sh *Helper) CreateValidator(addr sdk.ValAddress, pk cryptotypes.PubKey, stakeAmount sdk.Int, ok bool) {
	coin := sdk.NewCoin(sh.Denom, stakeAmount)
	sh.createValidator(addr, pk, coin, ok)
}

// CreateValidatorWithValPower calls staking module `MsgServer/CreateValidator` to create a new validator with zero
// commission
func (sh *Helper) CreateValidatorWithValPower(addr sdk.ValAddress, pk cryptotypes.PubKey, valPower int64, ok bool) sdk.Int {
	amount := sh.k.TokensFromConsensusPower(sh.Ctx, valPower)
	coin := sdk.NewCoin(sh.Denom, amount)
	sh.createValidator(addr, pk, coin, ok)
	return amount
}

// CreateValidatorMsg returns a message used to create validator in this service.
func (sh *Helper) CreateValidatorMsg(addr sdk.ValAddress, pk cryptotypes.PubKey, stakeAmount sdk.Int) *stakingtypes.MsgCreateValidator {
	coin := sdk.NewCoin(sh.Denom, stakeAmount)
	msg, err := stakingtypes.NewMsgCreateValidator(addr, pk, coin, stakingtypes.Description{}, sh.Commission, sdk.OneInt())
	require.NoError(sh.t, err)
	return msg
}

func (sh *Helper) createValidator(addr sdk.ValAddress, pk cryptotypes.PubKey, coin sdk.Coin, ok bool) {
	msg, err := stakingtypes.NewMsgCreateValidator(addr, pk, coin, stakingtypes.Description{}, sh.Commission, sdk.OneInt())
	require.NoError(sh.t, err)
	res, err := sh.MsgSrvr.CreateValidator(sdk.WrapSDKContext(sh.Ctx), msg)
	if ok {
		require.NoError(sh.t, err)
		require.NotNil(sh.t, res)
	} else {
		require.Error(sh.t, err)
		require.Nil(sh.t, res)
	}
}

// Delegate calls staking module staking module `MsgServer/Delegate` to delegate stake for a validator
func (sh *Helper) Delegate(delegator sdk.AccAddress, val sdk.ValAddress, amount sdk.Int) {
	coin := sdk.NewCoin(sh.Denom, amount)
	msg := stakingtypes.NewMsgDelegate(delegator, val, coin)
	res, err := sh.MsgSrvr.Delegate(sdk.WrapSDKContext(sh.Ctx), msg)
	require.NoError(sh.t, err)
	require.NotNil(sh.t, res)
}

// DelegateWithPower calls staking module `MsgServer/Delegate` to delegate stake for a validator
func (sh *Helper) DelegateWithPower(delegator sdk.AccAddress, val sdk.ValAddress, power int64) {
	coin := sdk.NewCoin(sh.Denom, sh.k.TokensFromConsensusPower(sh.Ctx, power))
	msg := stakingtypes.NewMsgDelegate(delegator, val, coin)
	res, err := sh.MsgSrvr.Delegate(sdk.WrapSDKContext(sh.Ctx), msg)
	require.NoError(sh.t, err)
	require.NotNil(sh.t, res)
}

// Undelegate calls staking module `MsgServer/Undelegate` to unbound some stake from a validator.
func (sh *Helper) Undelegate(delegator sdk.AccAddress, val sdk.ValAddress, amount sdk.Int, ok bool) {
	unbondAmt := sdk.NewCoin(sh.Denom, amount)
	msg := stakingtypes.NewMsgUndelegate(delegator, val, unbondAmt)
	res, err := sh.MsgSrvr.Undelegate(sdk.WrapSDKContext(sh.Ctx), msg)
	if ok {
		require.NoError(sh.t, err)
		require.NotNil(sh.t, res)
	} else {
		require.Error(sh.t, err)
		require.Nil(sh.t, res)
	}
}

// CheckValidator asserts that a validor exists and has a given status (if status!="")
// and if has a right jailed flag.
func (sh *Helper) CheckValidator(addr sdk.ValAddress, status stakingtypes.BondStatus, jailed bool) stakingtypes.Validator {
	v, ok := sh.k.GetValidator(sh.Ctx, addr)
	require.True(sh.t, ok)
	require.Equal(sh.t, jailed, v.Jailed, "wrong Jalied status")
	if status >= 0 {
		require.Equal(sh.t, status, v.Status)
	}
	return v
}

// CheckDelegator asserts that a delegator exists
func (sh *Helper) CheckDelegator(delegator sdk.AccAddress, val sdk.ValAddress, found bool) {
	_, ok := sh.k.GetDelegation(sh.Ctx, delegator, val)
	require.Equal(sh.t, ok, found)
}

// TurnBlock calls EndBlocker and updates the block time
func (sh *Helper) TurnBlock(newTime time.Time) sdk.Context {
	sh.Ctx = sh.Ctx.WithBlockTime(newTime)
	staking.EndBlocker(sh.Ctx, sh.k)
	return sh.Ctx
}

// TurnBlockTimeDiff calls EndBlocker and updates the block time by adding the
// duration to the current block time
func (sh *Helper) TurnBlockTimeDiff(diff time.Duration) sdk.Context {
	sh.Ctx = sh.Ctx.WithBlockTime(sh.Ctx.BlockHeader().Time.Add(diff))
	staking.EndBlocker(sh.Ctx, sh.k)
	return sh.Ctx
}

// ZeroCommission constructs a commission rates with all zeros.
func ZeroCommission() stakingtypes.CommissionRates {
	return stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
}
