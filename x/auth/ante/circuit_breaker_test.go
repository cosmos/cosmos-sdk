package ante_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type mockBreaker map[string]bool

func (m mockBreaker) IsAllowed(_ context.Context, typeURL string) (bool, error) {
	return !m[typeURL], nil
}

type cbMockTx struct{ msgs []sdk.Msg }

func (t cbMockTx) GetMsgs() []sdk.Msg                    { return t.msgs }
func (t cbMockTx) GetMsgsV2() ([]protov2.Message, error) { return nil, nil }
func (t cbMockTx) ValidateBasic() error                  { return nil }
func (t cbMockTx) Hash() [32]byte                        { return [32]byte{} }
func (t cbMockTx) GetSigners() ([][]byte, error)         { return nil, nil }

var (
	grantee = sdk.AccAddress("grantee_____________")
	send    = &banktypes.MsgSend{}
	deleg   = &stakingtypes.MsgDelegate{}
)

func run(t *testing.T, breaker mockBreaker, msgs ...sdk.Msg) error {
	t.Helper()
	called := false
	next := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		called = true
		return ctx, nil
	}
	_, err := ante.NewCircuitBreakerDecorator(breaker).AnteHandle(sdk.Context{}, cbMockTx{msgs: msgs}, false, next)
	require.Equal(t, err == nil, called, "next must run only when the decorator passes")
	return err
}

func exec(msgs ...sdk.Msg) *authz.MsgExec {
	m := authz.NewMsgExec(grantee, msgs)
	return &m
}

func TestCircuitBreakerDecoratorTopLevel(t *testing.T) {
	breaker := mockBreaker{sdk.MsgTypeURL(send): true}

	require.NoError(t, run(t, breaker, deleg))
	require.ErrorIs(t, run(t, breaker, send), sdkerrors.ErrCircuitBreakerTripped)
	require.ErrorIs(t, run(t, breaker, deleg, send), sdkerrors.ErrCircuitBreakerTripped)
}

// In FinalizeBlock the decorator leaves the check to the router, which runs
// after the ante handler has taken the fee and used the sequence, so a tx a
// trip stops in a block cannot be replayed once the trip is reset.
func TestCircuitBreakerDecoratorDefersToRouterInFinalizeBlock(t *testing.T) {
	breaker := mockBreaker{sdk.MsgTypeURL(send): true}
	modes := []sdk.ExecMode{
		sdk.ExecModeCheck, sdk.ExecModeReCheck, sdk.ExecModeSimulate,
		sdk.ExecModePrepareProposal, sdk.ExecModeProcessProposal, sdk.ExecModeFinalize,
	}
	for _, mode := range modes {
		called := false
		next := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
			called = true
			return ctx, nil
		}
		ctx := sdk.Context{}.WithExecMode(mode)
		// baseapp passes simulate=true only in ExecModeSimulate
		simulate := mode == sdk.ExecModeSimulate
		_, err := ante.NewCircuitBreakerDecorator(breaker).AnteHandle(ctx, cbMockTx{msgs: []sdk.Msg{exec(send)}}, simulate, next)
		if mode == sdk.ExecModeFinalize {
			require.NoError(t, err, "mode %d", mode)
			require.True(t, called, "mode %d", mode)
			continue
		}
		require.ErrorIs(t, err, sdkerrors.ErrCircuitBreakerTripped, "mode %d", mode)
		require.False(t, called, "mode %d", mode)
	}
}

// contrib/x/circuit's decorator only checks top-level msgs.
func TestCircuitBreakerDecoratorNested(t *testing.T) {
	breaker := mockBreaker{sdk.MsgTypeURL(send): true}

	require.NoError(t, run(t, breaker, exec(deleg)))
	require.ErrorIs(t, run(t, breaker, exec(send)), sdkerrors.ErrCircuitBreakerTripped)
	require.ErrorIs(t, run(t, breaker, exec(deleg, send)), sdkerrors.ErrCircuitBreakerTripped)
	require.ErrorIs(t, run(t, breaker, exec(exec(exec(send)))), sdkerrors.ErrCircuitBreakerTripped)

	// a disabled msg after an allowed wrapper in the same list is still
	// checked, at the top level and inside another wrapper
	require.ErrorIs(t, run(t, breaker, exec(deleg), send), sdkerrors.ErrCircuitBreakerTripped)
	require.ErrorIs(t, run(t, breaker, exec(exec(deleg), send)), sdkerrors.ErrCircuitBreakerTripped)
	require.ErrorIs(t, run(t, breaker, proposal(t, exec(deleg), send)), sdkerrors.ErrCircuitBreakerTripped)

	// disabling MsgExec itself blocks the wrapper, at the top level and nested
	// inside another wrapper
	breaker = mockBreaker{sdk.MsgTypeURL(&authz.MsgExec{}): true}
	require.ErrorIs(t, run(t, breaker, exec(deleg)), sdkerrors.ErrCircuitBreakerTripped)
	require.ErrorIs(t, run(t, breaker, proposal(t, exec(deleg))), sdkerrors.ErrCircuitBreakerTripped)
	require.ErrorIs(t, run(t, breaker, proposal(t, proposal(t, exec(deleg)))), sdkerrors.ErrCircuitBreakerTripped)
}

func proposal(t *testing.T, msgs ...sdk.Msg) *govv1.MsgSubmitProposal {
	t.Helper()
	m, err := govv1.NewMsgSubmitProposal(msgs, nil, grantee.String(), "", "title", "summary", false)
	require.NoError(t, err)
	return m
}

// A wrapper exposing GetMsgs, such as gov v1 MsgSubmitProposal, is opened even
// though the router treats it as protected.
func TestCircuitBreakerDecoratorOpensGetMsgsWrappers(t *testing.T) {
	breaker := mockBreaker{sdk.MsgTypeURL(send): true}

	require.NoError(t, run(t, breaker, proposal(t, deleg)))
	require.NoError(t, run(t, breaker, proposal(t)))
	require.ErrorIs(t, run(t, breaker, proposal(t, send)), sdkerrors.ErrCircuitBreakerTripped)
	require.ErrorIs(t, run(t, breaker, proposal(t, deleg, send)), sdkerrors.ErrCircuitBreakerTripped)
	require.ErrorIs(t, run(t, breaker, proposal(t, exec(send))), sdkerrors.ErrCircuitBreakerTripped)
	require.ErrorIs(t, run(t, breaker, exec(proposal(t, send))), sdkerrors.ErrCircuitBreakerTripped)
}

func TestCircuitBreakerDecoratorDepthLimit(t *testing.T) {
	var msg sdk.Msg = deleg
	for range ante.MaxNestedMsgDepth {
		msg = exec(msg)
	}
	require.NoError(t, run(t, mockBreaker{}, msg))

	msg = exec(msg)
	require.ErrorIs(t, run(t, mockBreaker{}, msg), sdkerrors.ErrInvalidRequest)

	// the limit counts GetMsgs wrappers too
	msg = deleg
	for range ante.MaxNestedMsgDepth {
		msg = proposal(t, msg)
	}
	require.NoError(t, run(t, mockBreaker{}, msg))
	require.ErrorIs(t, run(t, mockBreaker{}, proposal(t, msg)), sdkerrors.ErrInvalidRequest)
}

type errBreaker struct {
	allowed bool
	err     error
}

func (b errBreaker) IsAllowed(context.Context, string) (bool, error) { return b.allowed, b.err }

// A breaker error rejects the tx with that error, whether the breaker reports
// allowed or not: the decorator fails closed and returns the breaker's error,
// such as a store error or the keeper's error naming the blocking pattern.
func TestCircuitBreakerDecoratorFailsClosedOnBreakerError(t *testing.T) {
	boom := errors.New("store unavailable")
	for _, allowed := range []bool{true, false} {
		called := false
		next := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
			called = true
			return ctx, nil
		}
		_, err := ante.NewCircuitBreakerDecorator(errBreaker{allowed, boom}).AnteHandle(sdk.Context{}, cbMockTx{msgs: []sdk.Msg{deleg}}, false, next)
		require.ErrorIs(t, err, boom, "allowed=%t", allowed)
		require.False(t, called, "allowed=%t", allowed)
	}
}

var errUnpack = errors.New("cannot unpack inner msgs")

// badGetMessages and badGetMsgs are wrappers whose inner msgs cannot be read.
type (
	badGetMessages struct{ *banktypes.MsgSend }
	badGetMsgs     struct{ *banktypes.MsgSend }
)

func (badGetMessages) GetMessages() ([]sdk.Msg, error) { return nil, errUnpack }
func (badGetMsgs) GetMsgs() ([]sdk.Msg, error)         { return nil, errUnpack }

// A wrapper the decorator cannot open is rejected, since it might wrap a
// disabled msg.
func TestCircuitBreakerDecoratorRejectsUnopenableWrapper(t *testing.T) {
	for _, msg := range []sdk.Msg{badGetMessages{send}, badGetMsgs{send}, exec(badGetMsgs{send})} {
		require.ErrorIs(t, run(t, mockBreaker{}, msg), errUnpack)
	}
}

// bothWrapper exposes msgs through both GetMessages and GetMsgs.
type bothWrapper struct {
	*stakingtypes.MsgDelegate
	messages, msgs []sdk.Msg
}

func (w bothWrapper) GetMessages() ([]sdk.Msg, error) { return w.messages, nil }
func (w bothWrapper) GetMsgs() ([]sdk.Msg, error)     { return w.msgs, nil }

// A wrapper exposing both methods has both lists checked, so a disabled msg
// cannot hide in either one.
func TestCircuitBreakerDecoratorOpensBothWrapperLists(t *testing.T) {
	breaker := mockBreaker{sdk.MsgTypeURL(send): true}

	require.NoError(t, run(t, breaker, bothWrapper{deleg, []sdk.Msg{deleg}, []sdk.Msg{deleg}}))
	require.ErrorIs(t, run(t, breaker, bothWrapper{deleg, []sdk.Msg{send}, nil}), sdkerrors.ErrCircuitBreakerTripped)
	require.ErrorIs(t, run(t, breaker, bothWrapper{deleg, nil, []sdk.Msg{send}}), sdkerrors.ErrCircuitBreakerTripped)
	require.ErrorIs(t, run(t, breaker, bothWrapper{deleg, []sdk.Msg{deleg}, []sdk.Msg{send}}), sdkerrors.ErrCircuitBreakerTripped)
	require.ErrorIs(t, run(t, breaker, bothWrapper{deleg, []sdk.Msg{send}, []sdk.Msg{deleg}}), sdkerrors.ErrCircuitBreakerTripped)
}

// TestNewAnteHandlerCircuitBreaker sets HandlerOptions.CircuitBreaker and
// sends an unsigned tx. A tripped msg fails at the breaker, before fee
// deduction calls the bank keeper mock; an allowed msg passes the breaker
// and fails later, at signature checks.
func TestNewAnteHandlerCircuitBreaker(t *testing.T) {
	suite := SetupTestSuite(t, true)
	addr := sdk.AccAddress("sender______________")
	msg := testdata.NewTestMsg(addr)

	for _, tc := range []struct {
		name    string
		breaker mockBreaker
		wantErr error
	}{
		{"tripped", mockBreaker{sdk.MsgTypeURL(msg): true}, sdkerrors.ErrCircuitBreakerTripped},
		{"allowed", mockBreaker{}, sdkerrors.ErrNoSignatures},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handler, err := ante.NewAnteHandler(ante.HandlerOptions{
				AccountKeeper:   suite.accountKeeper,
				BankKeeper:      suite.bankKeeper,
				SignModeHandler: suite.encCfg.TxConfig.SignModeHandler(),
				CircuitBreaker:  tc.breaker,
			})
			require.NoError(t, err)

			builder := suite.clientCtx.TxConfig.NewTxBuilder()
			require.NoError(t, builder.SetMsgs(msg))
			builder.SetGasLimit(testdata.NewTestGasLimit())
			builder.SetFeeAmount(testdata.NewTestFeeAmount())
			_, err = handler(suite.ctx, builder.GetTx(), false)
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}
