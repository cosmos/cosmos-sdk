package safety_test

import (
	"context"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	safetytypes "github.com/cosmos/cosmos-sdk/x/safety/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// TestUpgradeHandlerRunsUnderTrip shows that a governance-approved upgrade
// applies under a "*" trip even when its handler calls a msg server directly,
// as handlers often call bank UpdateParams, or sends msgs through the
// MsgServiceRouter or a hybrid handler, while a direct call made in the same
// PreBlocker outside the handler, one in the next BeginBlocker, and a routed
// msg in a later tx are still refused.
func TestUpgradeHandlerRunsUnderTrip(t *testing.T) {
	const name = "incident-fix"
	f, p := newPrecompileFixture(t, false)

	f.mustDeliver(f.operator, safetytypes.NewMsgTrip(f.operator.String(), safetytypes.Wildcard))
	// the proposal takes three blocks: submit, vote, and the tally
	height := f.app.LastBlockHeight() + 10
	f.mustPass(&upgradetypes.MsgSoftwareUpgrade{Authority: govAddr, Plan: upgradetypes.Plan{Name: name, Height: height}})
	p.openProposal()
	require.Less(t, f.app.LastBlockHeight(), height-1)
	for f.app.LastBlockHeight() < height-1 {
		f.finalize()
	}
	// a binary with the handler before the upgrade height halts the chain,
	// so register it now, as the upgraded binary would
	var (
		ran    bool
		bypass bool
	)
	voterBefore := f.balance(f.voter.addr)
	f.app.UpgradeKeeper.SetUpgradeHandler(name, func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ran = true
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		bypass = sdkCtx.CircuitBreakerBypass()
		routed := send(f.user, f.voter)
		if _, err := f.app.MsgServiceRouter().Handler(routed)(sdkCtx, routed); err != nil {
			return vm, err
		}
		hybrid := f.app.MsgServiceRouter().HybridHandlerByMsgName(proto.MessageName(routed))
		if err := hybrid(ctx, send(f.user, f.voter), &banktypes.MsgSendResponse{}); err != nil {
			return vm, err
		}
		params := banktypes.DefaultParams()
		params.DefaultSendEnabled = false
		_, err := bankkeeper.NewMsgServerImpl(f.app.BankKeeper).UpdateParams(ctx, &banktypes.MsgUpdateParams{Authority: govAddr, Params: params})
		return vm, err
	})

	// the upgrade block: the precompile runs in the PreBlocker before the
	// upgrade module applies the plan
	p.runAt(t, sitePreBlocker, msgSendURL, msgDelegateURL, msgWithdrawRewardURL)
	require.True(t, p.breaker)
	require.True(t, ran, "the upgrade handler did not run")
	require.True(t, bypass, "the upgrade handler Context does not bypass the circuit breaker")

	done, err := f.app.UpgradeKeeper.GetDoneHeight(f.ctx(), name)
	require.NoError(t, err)
	require.Equal(t, height, done)
	require.False(t, f.app.BankKeeper.GetParams(f.ctx()).DefaultSendEnabled, "the handler's UpdateParams did not apply")
	require.Equal(t, voterBefore.AddRaw(2), f.balance(f.voter.addr), "the handler's routed sends did not apply")

	p.runAt(t, siteBeginBlocker, msgSendURL, msgDelegateURL, msgWithdrawRewardURL)
	require.True(t, p.breaker)
	f.requireFails(f.user, safetytypes.ErrMsgDisabled, send(f.user, f.voter))
}
