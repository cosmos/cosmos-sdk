package staking_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
	"time"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

type trace struct {
	_Meta  interface{} `json:"#meta"`
	States []struct {
		Action struct {
			Amount       int    `json:"amount"`
			Delegator    string `json:"delegator"`
			HeightDelta  int    `json:"heightDelta"`
			Nature       string `json:"nature"`
			TimeDelta    int    `json:"timeDelta"`
			Validator    string `json:"validator"`
			ValidatorDst string `json:"validatorDst"`
			ValidatorSrc string `json:"validatorSrc"`
		} `json:"action"`
		BlockHeight int `json:"blockHeight"`
		BlockTime   int `json:"blockTime"`
		Delegation  struct {
			_Map [][]interface{} `json:"#map"`
		} `json:"delegation"`
		Jailed struct {
			V0 bool `json:"v0"`
			V1 bool `json:"v1"`
		} `json:"jailed"`
		LastValSet struct {
			_Set []string `json:"#set"`
		} `json:"lastValSet"`
		Outcome       string `json:"outcome"`
		RedelegationQ struct {
			_Set []struct {
				CompletionTime int    `json:"completionTime"`
				CreationHeight int    `json:"creationHeight"`
				Delegator      string `json:"delegator"`
				InitialBalance int    `json:"initialBalance"`
				SharesDst      int    `json:"sharesDst"`
				ValidatorDst   string `json:"validatorDst"`
				ValidatorSrc   string `json:"validatorSrc"`
			} `json:"#set"`
		} `json:"redelegationQ"`
		Status struct {
			V0 string `json:"v0"`
			V1 string `json:"v1"`
		} `json:"status"`
		Steps  int `json:"steps"`
		Tokens struct {
			D0 int `json:"d0"`
			V0 int `json:"v0"`
			V1 int `json:"v1"`
		} `json:"tokens"`
		UnbondingHeight struct {
			V0 int `json:"v0"`
			V1 int `json:"v1"`
		} `json:"unbondingHeight"`
		UnbondingTime struct {
			V0 int `json:"v0"`
			V1 int `json:"v1"`
		} `json:"unbondingTime"`
		UndelegationQ struct {
			_Set []struct {
				Balance        int    `json:"balance"`
				CompletionTime int    `json:"completionTime"`
				CreationHeight int    `json:"creationHeight"`
				Delegator      string `json:"delegator"`
				Validator      string `json:"validator"`
			} `json:"#set"`
		} `json:"undelegationQ"`
		ValidatorQ struct {
			_Set []string `json:"#set"`
		} `json:"validatorQ"`
	} `json:"states"`
	Vars []string `json:"vars"`
}

func loadTraces(fn string) []trace {

	fd, err := os.Open(fn)

	if err != nil {
		panic(err)
	}

	defer fd.Close()

	byteValue, _ := ioutil.ReadAll(fd)

	var ret []trace

	json.Unmarshal([]byte(byteValue), &ret)

	return ret
}

func scaledAmt(modelAmt int) sdk.Int {
	return sdk.TokensFromConsensusPower(int64(modelAmt), sdk.DefaultPowerReduction)
}

type helper struct {
	t          *testing.T
	h          sdk.Handler
	k          keeper.Keeper
	bk         types.BankKeeper
	ctx        sdk.Context
	commission stakingtypes.CommissionRates
	denom      string
}

// constructs a commission rates with all zeros.
func zeroCommission() stakingtypes.CommissionRates {
	return stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
}

// creates staking Handler wrapper for tests
func newHelper(t *testing.T, ctx sdk.Context, k keeper.Keeper, bk types.BankKeeper) *helper {
	return &helper{t, staking.NewHandler(k), k, bk, ctx, zeroCommission(), sdk.DefaultBondDenom}
}

// calls handler to create a new staking validator
func (h *helper) createValidator(addr sdk.ValAddress, pk cryptotypes.PubKey, stakeAmt int, ok bool) {
	coin := sdk.NewCoin(h.denom, scaledAmt(stakeAmt))
	msg, err := stakingtypes.NewMsgCreateValidator(addr, pk, coin, stakingtypes.Description{}, h.commission, sdk.OneInt())
	require.NoError(h.t, err)
	h.handle(msg, ok)
}

// calls handler to delegate stake for a validator
func (h *helper) delegate(delegator sdk.AccAddress, val sdk.ValAddress, amt int, ok bool) {
	coin := sdk.NewCoin(h.denom, scaledAmt(amt))
	msg := stakingtypes.NewMsgDelegate(delegator, val, coin)
	h.handle(msg, ok)
}

// calls handler to unbound some stake from a validator.
func (h *helper) undelegate(delegator sdk.AccAddress, val sdk.ValAddress, amt int, ok bool) *sdk.Result {
	coin := sdk.NewCoin(h.denom, scaledAmt(amt))
	msg := stakingtypes.NewMsgUndelegate(delegator, val, coin)
	return h.handle(msg, ok)
}

// calls handler to redelegate funds from src to dst
func (h *helper) beginRedelegate(delegator sdk.AccAddress, src sdk.ValAddress, dst sdk.ValAddress, amt int, ok bool) *sdk.Result {
	coin := sdk.NewCoin(h.denom, scaledAmt(amt))
	msg := stakingtypes.NewMsgBeginRedelegate(delegator, src, dst, coin)
	return h.handle(msg, ok)
}

func (h *helper) matchValidatorStatus(val sdk.ValAddress, status string) {
	validator, _ := h.k.GetValidator(h.ctx, val)
	actual := validator.GetStatus()
	if status == "bonded" {
		require.Equal(h.t, types.Bonded, actual)
	}
	if status == "unbonding" {
		require.Equal(h.t, types.Unbonding, actual)
	}
	if status == "unbonded" {
		require.Equal(h.t, types.Unbonded, actual)
	}
}

func (h *helper) matchBalance(addr sdk.AccAddress, amt int) {
	bal := h.bk.GetBalance(h.ctx, addr, h.denom)
	exp := sdk.NewCoin(h.denom, scaledAmt(amt))
	require.Equal(h.t, exp, bal)
}

func (h *helper) matchTokens(addr sdk.ValAddress, amt int) {
	validator, _ := h.k.GetValidator(h.ctx, addr)
	tok := validator.Tokens
	exp := scaledAmt(amt)
	require.Equal(h.t, exp, tok)
}

func (h *helper) ensureValidatorLexicographicOrderingMatchesModel(v0 sdk.ValAddress, v1 sdk.ValAddress) {
	/*
		Ties in validator power are broken based on comparing PowerIndexKey. The model tie-break needs
		to match the code tie-break at all times. This function ensures the tie break function in the model
		is correct.
	*/
	xv, _ := h.k.GetValidator(h.ctx, v0)
	yv, _ := h.k.GetValidator(h.ctx, v1)
	xk := types.GetValidatorsByPowerIndexKey(xv, sdk.DefaultPowerReduction)
	yk := types.GetValidatorsByPowerIndexKey(yv, sdk.DefaultPowerReduction)
	// The result will be 0 if a==b, -1 if a < b, and +1 if a > b.
	res := bytes.Compare(xk, yk)
	// Confirm that validator precedence is the same in code as in model
	require.Equal(h.t, 1, res)
}

// handle calls staking handler on a given message
func (h *helper) handle(msg sdk.Msg, ok bool) *sdk.Result {
	res, err := h.h(h.ctx, msg)
	if ok {
		require.NoError(h.t, err)
		require.NotNil(h.t, res)
	} else {
		require.Error(h.t, err)
		require.Nil(h.t, res)
	}
	return res
}

func ExecuteTrace(t *testing.T, trace trace) {
	ix := make(map[string]int)
	ix["v0"] = 0
	ix["v1"] = 1
	ix["d0"] = 4

	initPower := int64(0)
	numAddresses := 6
	app, ctx, delAddrs, valAddrs := bootstrapHandlerGenesisTest(t, initPower, numAddresses, sdk.TokensFromConsensusPower(initPower, sdk.DefaultPowerReduction))
	validators, delegators := valAddrs[:3], delAddrs

	params := app.StakingKeeper.GetParams(ctx)
	params.UnbondingTime = 2 * time.Second
	params.MaxValidators = 1
	app.StakingKeeper.SetParams(ctx, params)

	h := newHelper(t, ctx, app.StakingKeeper, app.BankKeeper)

	var states = trace.States
	var init = states[0]
	states = states[1:]

	require.NoError(t, simapp.FundAccount(app.BankKeeper, ctx, delegators[ix["v0"]], sdk.NewCoins(sdk.NewCoin(params.BondDenom, scaledAmt(init.Tokens.V0)))))
	require.NoError(t, simapp.FundAccount(app.BankKeeper, ctx, delegators[ix["v1"]], sdk.NewCoins(sdk.NewCoin(params.BondDenom, scaledAmt(init.Tokens.V1)))))
	require.NoError(t, simapp.FundAccount(app.BankKeeper, ctx, delegators[ix["d0"]], sdk.NewCoins(sdk.NewCoin(params.BondDenom, scaledAmt(init.Tokens.D0)))))

	h.createValidator(validators[0], PKs[0], 1, true)
	h.createValidator(validators[1], PKs[1], 1, true)

	h.ensureValidatorLexicographicOrderingMatchesModel(validators[ix["v0"]], validators[ix["v1"]])

	h.matchValidatorStatus(validators[ix["v0"]], init.Status.V0)
	h.matchValidatorStatus(validators[ix["v1"]], init.Status.V1)

	for _, state := range states {
		var succeed = state.Outcome == "succeed"
		switch state.Action.Nature {
		case "endBlock":
			// Does this make sense?
			var dt = time.Duration(int64(state.Action.TimeDelta) * int64(time.Second))
			staking.EndBlocker(h.ctx, h.k)
			h.ctx = h.ctx.WithBlockTime(h.ctx.BlockHeader().Time.Add(dt))
		case "delegate":
			var del = delegators[ix[state.Action.Delegator]]
			var val = validators[ix[state.Action.Validator]]
			h.delegate(del, val, state.Action.Amount, succeed)
		case "undelegate":
			var del = delegators[ix[state.Action.Delegator]]
			var val = validators[ix[state.Action.Validator]]
			h.undelegate(del, val, state.Action.Amount, succeed)
		case "beginRedelegate":
			var del = delegators[ix[state.Action.Delegator]]
			var src = validators[ix[state.Action.ValidatorSrc]]
			var dst = validators[ix[state.Action.ValidatorDst]]
			h.beginRedelegate(del, src, dst, state.Action.Amount, succeed)
		}
		h.matchValidatorStatus(validators[ix["v0"]], state.Status.V0)
		h.matchValidatorStatus(validators[ix["v1"]], state.Status.V1)
		h.matchBalance(delegators[ix["d0"]], state.Tokens.D0)
		h.matchTokens(validators[ix["v0"]], state.Tokens.V0)
		h.matchTokens(validators[ix["v1"]], state.Tokens.V1)
	}
}

func ExecuteTraces(t *testing.T, traces []trace) {
	for _, trace := range traces {
		ExecuteTrace(t, trace)
	}
}

func TestTracesQueues(t *testing.T) {
	// Test several traces chosen by projecting model states to
	// (sign(validatorQ.size,) sign(undelegationQ.size), sign(redelegationQ.size))
	ExecuteTraces(t, loadTraces("mbt/model_based_testing_traces_queues.json"))
}
func TestTracesActionOutcome(t *testing.T) {
	// Test several traces chosen by projecting model states to
	// (action.nature, outcome)
	ExecuteTraces(t, loadTraces("mbt/model_based_testing_traces_action_outcome.json"))
}
func TestTracesP0(t *testing.T) {
	// Test a trace matching predicate P<n>. Please see mbt/staking.tla
	ExecuteTraces(t, loadTraces("mbt/model_based_testing_traces_P0.json"))
}
func TestTracesP1(t *testing.T) {
	// Test a trace matching predicate P<n>. Please see mbt/staking.tla
	ExecuteTraces(t, loadTraces("mbt/model_based_testing_traces_P1.json"))
}
func TestTracesP2(t *testing.T) {
	// Test a trace matching predicate P<n>. Please see mbt/staking.tla
	ExecuteTraces(t, loadTraces("mbt/model_based_testing_traces_P2.json"))
}
func TestTracesP3(t *testing.T) {
	// Test a trace matching predicate P<n>. Please see mbt/staking.tla
	ExecuteTraces(t, loadTraces("mbt/model_based_testing_traces_P3.json"))
}
func TestTracesP4(t *testing.T) {
	// Test a trace matching predicate P<n>. Please see mbt/staking.tla
	ExecuteTraces(t, loadTraces("mbt/model_based_testing_traces_P4.json"))
}
func TestTracesP5(t *testing.T) {
	// Test a trace matching predicate P<n>. Please see mbt/staking.tla
	ExecuteTraces(t, loadTraces("mbt/model_based_testing_traces_P5.json"))
}
func TestTracesP6(t *testing.T) {
	// Test a trace matching predicate P<n>. Please see mbt/staking.tla
	ExecuteTraces(t, loadTraces("mbt/model_based_testing_traces_P6.json"))
}
func TestTracesP7(t *testing.T) {
	// Test a trace matching predicate P<n>. Please see mbt/staking.tla
	ExecuteTraces(t, loadTraces("mbt/model_based_testing_traces_P7.json"))
}
func TestTracesP8(t *testing.T) {
	// Test a trace matching predicate P<n>. Please see mbt/staking.tla
	ExecuteTraces(t, loadTraces("mbt/model_based_testing_traces_P8.json"))
}
