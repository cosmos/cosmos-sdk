package safety_test

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/simapp"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	safetytypes "github.com/cosmos/cosmos-sdk/x/safety/types"
)

// allowAll is a breaker that allows every msg without touching state.
type allowAll struct{}

func (allowAll) IsAllowed(context.Context, string) (bool, error) { return true, nil }

// newBaselineFixture starts a SimApp with no circuit breaker on the router
// and a CircuitBreakerDecorator that reads nothing, so its gas is what the same txs
// cost without x/safety.
func newBaselineFixture(t *testing.T) *fixture {
	t.Helper()
	return newFixtureWithSetup(t, nil, nil, func(app *simapp.SimApp) {
		app.SetCircuitBreaker(nil)
		anteHandler, err := ante.NewAnteHandler(ante.HandlerOptions{
			AccountKeeper:   app.AccountKeeper,
			BankKeeper:      app.BankKeeper,
			SignModeHandler: app.TxConfig().SignModeHandler(),
			FeegrantKeeper:  app.FeeGrantKeeper,
			SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
			SigVerifyOptions: []ante.SigVerificationDecoratorOption{
				ante.WithUnorderedTxGasCost(ante.DefaultUnorderedTxGasCost),
				ante.WithMaxUnorderedTxTimeoutDuration(ante.DefaultMaxTimeoutDuration),
			},
			CircuitBreaker: allowAll{},
		})
		require.NoError(t, err)
		app.SetAnteHandler(anteHandler)
	})
}

// withDisabled sets the genesis disabled list.
func withDisabled(t *testing.T, patterns []string) func(simapp.GenesisState) {
	t.Helper()
	return func(gs simapp.GenesisState) {
		var sgs safetytypes.GenesisState
		require.NoError(t, json.Unmarshal(gs[safetytypes.ModuleName], &sgs))
		sgs.DisabledPatterns = patterns
		bz, err := json.Marshal(&sgs)
		require.NoError(t, err)
		gs[safetytypes.ModuleName] = bz
	}
}

// txGas is the gas a tx used in Simulate and in FinalizeBlock.
type txGas struct {
	simulated, delivered uint64
}

// measure simulates a tx with msgs from signer, then delivers the same bytes
// and requires both to succeed.
func (f *fixture) measure(signer account, msgs ...sdk.Msg) txGas {
	f.t.Helper()
	bz := f.encode(signer, msgs...)
	info, _, err := f.app.Simulate(bz)
	require.NoError(f.t, err)
	res := f.finalize(bz).TxResults[0]
	require.Zero(f.t, res.Code, res.Log)
	return txGas{simulated: info.GasUsed, delivered: uint64(res.GasUsed)}
}

// gasCases runs a bank send, a staking delegation, an authz exec of a send,
// and a gov proposal of a send on f, in that order, and returns their gas.
func gasCases(t *testing.T, f *fixture) map[string]txGas {
	t.Helper()
	exp := f.now.Add(365 * 24 * time.Hour)
	grant, err := authz.NewMsgGrant(f.user.addr, f.operator.addr, authz.NewGenericAuthorization(msgSendURL), &exp)
	require.NoError(t, err)
	f.mustDeliver(f.user, grant)

	govParams, err := f.app.GovKeeper.Params.Get(f.ctx())
	require.NoError(t, err)
	govSend := send(account{addr: sdk.MustAccAddressFromBech32(govAddr)}, f.user)
	submit, err := govv1.NewMsgSubmitProposal([]sdk.Msg{govSend}, govParams.MinDeposit, f.voter.String(), "", "gas", "gas", false)
	require.NoError(t, err)

	valAddr := sdk.ValAddress(f.valSet.Validators[0].Address).String()
	exec := authz.NewMsgExec(f.operator.addr, []sdk.Msg{send(f.user, f.other)})
	gas := map[string]txGas{}
	gas["bank MsgSend"] = f.measure(f.user, send(f.user, f.other))
	gas["staking MsgDelegate"] = f.measure(f.user, stakingDelegate(f.user, valAddr))
	gas["authz MsgExec{MsgSend}"] = f.measure(f.operator, &exec)
	gas["gov MsgSubmitProposal{MsgSend}"] = f.measure(f.voter, submit)
	return gas
}

// TestBreakerCheckChargesNoGas checks that txs cost the same gas, both
// simulated and delivered, with x/safety wired as without a circuit breaker,
// with an empty disabled list and with 1000 entries that block none of them.
// The gap between simulated and delivered gas is the baseline's gap.
func TestBreakerCheckChargesNoGas(t *testing.T) {
	baseline := gasCases(t, newBaselineFixture(t))

	var many []string
	for i := range 1000 {
		many = append(many, fmt.Sprintf("/zz.m%04d.MsgX", i))
	}
	for name, disabled := range map[string][]string{"empty list": nil, "1000 entries": many} {
		t.Run(name, func(t *testing.T) {
			got := gasCases(t, newFixture(t, nil, withDisabled(t, disabled)))
			for _, tx := range slices.Sorted(maps.Keys(baseline)) {
				want := baseline[tx]
				t.Logf("%s: baseline sim %d deliver %d, wired sim %d deliver %d", tx, want.simulated, want.delivered, got[tx].simulated, got[tx].delivered)
				assert.Equal(t, want.delivered, got[tx].delivered, "%s delivered gas", tx)
				assert.Equal(t, want.simulated, got[tx].simulated, "%s simulated gas", tx)
				assert.Equal(t, int64(want.simulated)-int64(want.delivered), int64(got[tx].simulated)-int64(got[tx].delivered), "%s simulate gap", tx)
			}
		})
	}
}

// TestTrippedMsgStillRefused checks that, with the check charging no gas, a
// disabled msg still fails CheckTx and FinalizeBlock with the x/safety error.
func TestTrippedMsgStillRefused(t *testing.T) {
	f := newFixture(t, nil, withDisabled(t, []string{msgSendURL}))
	bz := f.requireCheckTxRejected(f.user, safetytypes.ErrMsgDisabled, send(f.user, f.other))
	res := f.finalize(bz).TxResults[0]
	require.Equal(t, safetytypes.ErrMsgDisabled.ABCICode(), res.Code, res.Log)
	require.Equal(t, safetytypes.ErrMsgDisabled.Codespace(), res.Codespace)
	require.Contains(t, res.Log, msgSendURL+" by "+msgSendURL+": msg type disabled by x/safety")
}
