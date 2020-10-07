package teststaking

import (
	"testing"

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
	h sdk.Handler

	Ctx        sdk.Context
	Commission stakingtypes.CommissionRates
	// Coin Denomination
	Denom string
}

// NewService creates staking Handler wrapper for tests
func NewService(ctx sdk.Context, k keeper.Keeper) *Service {
	return &Service{staking.NewHandler(k), ctx, ZeroCommission(), sdk.DefaultBondDenom}
}

// CreateValidator calls handler to create a new staking validator
func (sh *Service) CreateValidator(t *testing.T, addr sdk.ValAddress, pk crypto.PubKey, stakeAmount int64, ok bool) {
	coin := sdk.NewCoin(sh.Denom, sdk.NewInt(stakeAmount))
	sh.createValidator(t, addr, pk, coin, ok)
}

// CreateValidatorWithValPower calls handler to create a new staking validator with zero
// commission
func (sh *Service) CreateValidatorWithValPower(t *testing.T, addr sdk.ValAddress, pk crypto.PubKey, valPower int64, ok bool) sdk.Int {
	amount := sdk.TokensFromConsensusPower(valPower)
	coin := sdk.NewCoin(sh.Denom, amount)
	sh.createValidator(t, addr, pk, coin, ok)
	return amount
}

// CreateValidatorMsg returns a message used to create validator in this service.
func (sh *Service) CreateValidatorMsg(t *testing.T, addr sdk.ValAddress, pk crypto.PubKey, stakeAmount int64) *stakingtypes.MsgCreateValidator {
	coin := sdk.NewCoin(sh.Denom, sdk.NewInt(stakeAmount))
	msg, err := stakingtypes.NewMsgCreateValidator(addr, pk, coin, stakingtypes.Description{}, sh.Commission, sdk.OneInt())
	require.NoError(t, err)
	return msg
}

func (sh *Service) createValidator(t *testing.T, addr sdk.ValAddress, pk crypto.PubKey, coin sdk.Coin, ok bool) {
	msg, err := stakingtypes.NewMsgCreateValidator(addr, pk, coin, stakingtypes.Description{}, sh.Commission, sdk.OneInt())
	require.NoError(t, err)
	sh.Handle(t, msg, ok)
}

// Delegate calls handler to delegate stake for a validator
func (sh *Service) Delegate(t *testing.T, delAddr sdk.AccAddress, valAddr sdk.ValAddress, amount int64) {
	coin := sdk.NewCoin(sh.Denom, sdk.NewInt(amount))
	msg := stakingtypes.NewMsgDelegate(delAddr, valAddr, coin)
	sh.Handle(t, msg, true)
}

// DelegateWithPower calls handler to delegate stake for a validator
func (sh *Service) DelegateWithPower(t *testing.T, delAddr sdk.AccAddress, valAddr sdk.ValAddress, power int64) {
	coin := sdk.NewCoin(sh.Denom, sdk.TokensFromConsensusPower(power))
	msg := stakingtypes.NewMsgDelegate(delAddr, valAddr, coin)
	sh.Handle(t, msg, true)
}

// Undelegate calls handler to unbound some stake from a validator.
func (sh *Service) Undelegate(t *testing.T, delegator sdk.AccAddress, val sdk.ValAddress, amount sdk.Int, ok bool) *sdk.Result {
	unbondAmt := sdk.NewCoin(sh.Denom, amount)
	msg := stakingtypes.NewMsgUndelegate(delegator, val, unbondAmt)
	return sh.Handle(t, msg, ok)
}

// Handle calls staking handler on a given message
func (sh *Service) Handle(t *testing.T, msg sdk.Msg, ok bool) *sdk.Result {
	res, err := sh.h(sh.Ctx, msg)
	if ok {
		require.NoError(t, err)
		require.NotNil(t, res)
	} else {
		require.Error(t, err)
		require.Nil(t, res)
	}
	return res
}

// ZeroCommission constructs a commission rates with all zeros.
func ZeroCommission() stakingtypes.CommissionRates {
	return stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
}
