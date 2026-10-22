package safety_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/safety/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// branch returns a cached branch of ctx with its own gas meter and event
// manager.
func branch(ctx sdk.Context) sdk.Context {
	cacheCtx, _ := ctx.CacheContext()
	return cacheCtx.WithGasMeter(storetypes.NewInfiniteGasMeter()).WithEventManager(sdk.NewEventManager())
}

func (f *routerFixture) dispatchSend(ctx sdk.Context) (*banktypes.MsgSendResponse, error) {
	return msgservice.Dispatch[*banktypes.MsgSendResponse](ctx, f.app.MsgServiceRouter(), f.msgSend())
}

// TestDispatchBlockedByTrips dispatches a MsgSend under an exact, a prefix
// and a "*" trip. Each fails with the x/safety error, while a direct call
// to the bank msg server under the same trip moves the funds.
func TestDispatchBlockedByTrips(t *testing.T) {
	for _, pattern := range []string{urlSend, "/cosmos.bank.*", types.Wildcard} {
		t.Run(pattern, func(t *testing.T) {
			f := newRouterFixture(t, false)
			f.trip(pattern)

			resp, err := f.dispatchSend(branch(f.ctx()))
			require.ErrorIs(t, err, types.ErrMsgDisabled)
			require.ErrorContains(t, err, urlSend)
			require.Nil(t, resp)

			// a direct msg server call skips the router, so the trip does
			// not stop it
			ctx := branch(f.ctx())
			before := f.balance(ctx, f.grantee)
			_, err = bankkeeper.NewMsgServerImpl(f.bankKeeper).Send(ctx, f.msgSend().(*banktypes.MsgSend))
			require.NoError(t, err)
			require.Equal(t, before.AddRaw(1), f.balance(ctx, f.grantee))
		})
	}
}

// TestDispatchPrefixTripLeavesOtherMsgs dispatches a MsgDelegate under a
// bank prefix trip, which does not cover it.
func TestDispatchPrefixTripLeavesOtherMsgs(t *testing.T) {
	f := newRouterFixture(t, false)
	f.trip("/cosmos.bank.*")

	resp, err := msgservice.Dispatch[*stakingtypes.MsgDelegateResponse](branch(f.ctx()), f.app.MsgServiceRouter(), f.msgDelegate())
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestDispatchProtectedUnderWildcard dispatches a protected msg, the
// operator's reset, while "*" stands.
func TestDispatchProtectedUnderWildcard(t *testing.T) {
	f := newRouterFixture(t, false)
	f.trip(types.Wildcard)

	ctx := branch(f.ctx())
	resp, err := msgservice.Dispatch[*types.MsgResetResponse](ctx, f.app.MsgServiceRouter(), types.NewMsgReset(f.user.String(), types.Wildcard))
	require.NoError(t, err)
	require.NotNil(t, resp)

	has, err := f.safetyKeeper.DisabledList.Has(ctx, types.Wildcard)
	require.NoError(t, err)
	require.False(t, has)
}

// TestDispatchMatchesRouterPath runs a MsgSend through its router handler
// and through Dispatch on two branches of the same state. Both use the same
// gas, and Dispatch emits the events the router handler returns.
func TestDispatchMatchesRouterPath(t *testing.T) {
	f := newRouterFixture(t, false)
	ctx := f.ctx()

	routed := branch(ctx)
	res, err := f.app.MsgServiceRouter().Handler(f.msgSend())(routed, f.msgSend())
	require.NoError(t, err)
	require.NotEmpty(t, res.Events)

	dispatched := branch(ctx)
	resp, err := f.dispatchSend(dispatched)
	require.NoError(t, err)
	require.NotNil(t, resp)

	require.NotZero(t, routed.GasMeter().GasConsumed())
	require.Equal(t, routed.GasMeter().GasConsumed(), dispatched.GasMeter().GasConsumed())
	require.Equal(t, res.Events, dispatched.EventManager().ABCIEvents())
	require.Equal(t, f.balance(routed, f.grantee), f.balance(dispatched, f.grantee))
}

// TestDispatchGasMatchesNoBreaker dispatches an allowed MsgSend with x/safety
// as the breaker and runs the same msg through its router handler with no
// breaker set. The breaker check charges no gas, so both use the same gas.
func TestDispatchGasMatchesNoBreaker(t *testing.T) {
	f := newRouterFixture(t, false)
	ctx := f.ctx()

	dispatched := branch(ctx)
	_, err := f.dispatchSend(dispatched)
	require.NoError(t, err)

	router := f.app.MsgServiceRouter()
	router.SetCircuit(nil)
	t.Cleanup(func() { router.SetCircuit(f.safetyKeeper) })
	routed := branch(ctx)
	_, err = router.Handler(f.msgSend())(routed, f.msgSend())
	require.NoError(t, err)

	require.NotZero(t, routed.GasMeter().GasConsumed())
	require.Equal(t, routed.GasMeter().GasConsumed(), dispatched.GasMeter().GasConsumed())
}
