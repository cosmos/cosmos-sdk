package safety_test

import (
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	safetytypes "github.com/cosmos/cosmos-sdk/x/safety/types"
)

// noOverride returns the SimApp test consensus params without the
// AuthorityParams override, so every module falls back to its own authority.
func noOverride() *cmtproto.ConsensusParams {
	cp := defaultConsensusParams()
	cp.Authority = nil
	return cp
}

// TestAuthorityReplacesItself checks that, with no consensus override, the
// current authority hands the role to a new address through MsgUpdateParams,
// the new authority acts directly, the old one is locked out, and the new
// authority can hand the role on again.
func TestAuthorityReplacesItself(t *testing.T) {
	f := newFixture(t, noOverride(), nil)
	f.setParams(f.operator.String(), "/cosmos.bank.*")
	f.mustDeliver(f.operator, safetytypes.NewMsgTrip(f.operator.String(), msgSendURL))

	p := f.params()
	p.Authority = f.user.String()
	f.mustPass(&safetytypes.MsgUpdateParams{Authority: govAddr, Params: p})
	require.Equal(t, f.user.String(), f.params().Authority)
	require.Equal(t, []string{msgSendURL}, f.disabled(), "MsgUpdateParams must not change the disabled list")

	// gov is no longer the authority
	p.Scope = []string{"*"}
	reason := f.mustFailExecution(&safetytypes.MsgUpdateParams{Authority: govAddr, Params: p})
	require.Contains(t, reason, "invalid authority")
	reason = f.mustFailExecution(safetytypes.NewMsgReset(govAddr, msgSendURL))
	require.Contains(t, reason, "neither the authority nor the operator")
	require.Equal(t, []string{"/cosmos.bank.*"}, f.params().Scope)

	// the new authority resets and updates params in one tx each, no vote
	f.mustDeliver(f.user, safetytypes.NewMsgReset(f.user.String(), msgSendURL))
	require.Empty(t, f.disabled())
	f.mustDeliver(f.user, &safetytypes.MsgUpdateParams{Authority: f.user.String(), Params: p})
	require.Equal(t, []string{"*"}, f.params().Scope)

	// the new authority still cannot trip
	f.requireFails(f.user, sdkerrors.ErrUnauthorized, safetytypes.NewMsgTrip(f.user.String(), msgSendURL))

	// and hands the role on
	p.Authority = f.other.String()
	f.mustDeliver(f.user, &safetytypes.MsgUpdateParams{Authority: f.user.String(), Params: p})
	f.requireFails(f.user, sdkerrors.ErrUnauthorized, &safetytypes.MsgUpdateParams{Authority: f.user.String(), Params: p})
	require.Equal(t, f.other.String(), f.params().Authority)
}

// TestInvalidParamsRejected checks that the authority cannot store params
// that would break the module or genesis.
func TestInvalidParamsRejected(t *testing.T) {
	f := newFixture(t, nil, nil)
	good := f.params()

	for name, mutate := range map[string]func(*safetytypes.Params){
		"empty authority":   func(p *safetytypes.Params) { p.Authority = "" },
		"bad authority":     func(p *safetytypes.Params) { p.Authority = "cosmos1bad" },
		"bad operator":      func(p *safetytypes.Params) { p.Operator = "not-an-address" },
		"malformed scope":   func(p *safetytypes.Params) { p.Scope = []string{"cosmos.bank.*"} },
		"duplicate scope":   func(p *safetytypes.Params) { p.Scope = []string{"/cosmos.bank.*", "/cosmos.bank.*"} },
		"half wildcard":     func(p *safetytypes.Params) { p.Scope = []string{"/cosmos.bank*"} },
		"trailing dot only": func(p *safetytypes.Params) { p.Scope = []string{"/.*"} },
	} {
		t.Run(name, func(t *testing.T) {
			p := good
			p.Scope = append([]string(nil), good.Scope...)
			mutate(&p)
			reason := f.mustFailExecution(&safetytypes.MsgUpdateParams{Authority: govAddr, Params: p})
			require.NotEmpty(t, reason)
			require.Equal(t, good, f.params())
		})
	}
}

// TestConsensusAuthorityOverride checks that the consensus AuthorityParams
// override, once set, is the x/safety authority regardless of
// params.authority: the override address acts directly, and gov, still named
// in params.authority, is locked out.
func TestConsensusAuthorityOverride(t *testing.T) {
	f := newFixture(t, nil, nil)
	f.setParams(f.operator.String(), "/cosmos.bank.*")
	f.mustDeliver(f.operator, safetytypes.NewMsgTrip(f.operator.String(), msgSendURL))

	// while the override names gov, handing params.authority to another
	// address changes nothing: gov keeps the role
	p := f.params()
	p.Authority = f.other.String()
	f.mustPass(&safetytypes.MsgUpdateParams{Authority: govAddr, Params: p})
	require.Equal(t, f.other.String(), f.params().Authority)
	f.requireFails(f.other, sdkerrors.ErrUnauthorized, &safetytypes.MsgUpdateParams{Authority: f.other.String(), Params: p})
	f.requireFails(f.other, sdkerrors.ErrUnauthorized, safetytypes.NewMsgReset(f.other.String(), msgSendURL))
	p.Authority = govAddr
	f.mustPass(&safetytypes.MsgUpdateParams{Authority: govAddr, Params: p})

	// gov moves the consensus override to user
	cp, err := f.app.ConsensusParamsKeeper.ParamsStore.Get(f.ctx())
	require.NoError(t, err)
	f.mustPass(&consensustypes.MsgUpdateParams{
		Authority: govAddr,
		Block:     cp.Block,
		Evidence:  cp.Evidence,
		Validator: cp.Validator,
		Abci:      cp.Abci,
		Auth:      &cmtproto.AuthorityParams{Authority: f.user.String()},
	})
	require.Equal(t, govAddr, f.params().Authority, "params.authority is unchanged")

	// user is now the authority without being named in params
	f.mustDeliver(f.user, safetytypes.NewMsgReset(f.user.String(), msgSendURL))
	require.Empty(t, f.disabled())
	p.Scope = []string{"*"}
	f.mustDeliver(f.user, &safetytypes.MsgUpdateParams{Authority: f.user.String(), Params: p})
	require.Equal(t, []string{"*"}, f.params().Scope)

	// gov, still in params.authority, is not
	f.mustDeliver(f.operator, safetytypes.NewMsgTrip(f.operator.String(), msgSendURL))
	reason := f.mustFailExecution(safetytypes.NewMsgReset(govAddr, msgSendURL))
	require.Contains(t, reason, "neither the authority nor the operator")
	require.Equal(t, []string{msgSendURL}, f.disabled())

	// and the override address still cannot trip
	f.requireFails(f.user, sdkerrors.ErrUnauthorized, safetytypes.NewMsgTrip(f.user.String(), "/cosmos.bank.v1beta1.MsgMultiSend"))
}

// TestAuthorityResetsThroughAuthzUnderWildcard runs an authority that acts
// through an authz grantee. A "*" trip leaves authz MsgExec running, so the
// grantee resets "*" for the authority while MsgExec with an unprotected msg
// inside is still rejected. An exact trip of MsgExec closes that path until the
// operator resets it.
func TestAuthorityResetsThroughAuthzUnderWildcard(t *testing.T) {
	f := newFixture(t, noOverride(), nil)
	const execURL = "/cosmos.authz.v1beta1.MsgExec"

	// user is the authority, other its grantee
	p := f.params()
	p.Authority = f.user.String()
	p.Operator = f.operator.String()
	p.Scope = []string{safetytypes.Wildcard}
	f.mustPass(&safetytypes.MsgUpdateParams{Authority: govAddr, Params: p})
	for _, url := range []string{sdk.MsgTypeURL(&safetytypes.MsgReset{}), msgSendURL} {
		grant, err := authz.NewMsgGrant(f.user.addr, f.other.addr, authz.NewGenericAuthorization(url), nil)
		require.NoError(t, err)
		f.mustDeliver(f.user, grant)
	}
	exec := func(msg sdk.Msg) *authz.MsgExec {
		e := authz.NewMsgExec(f.other.addr, []sdk.Msg{msg})
		return &e
	}
	resetAll := safetytypes.NewMsgReset(f.user.String(), safetytypes.Wildcard)

	f.mustDeliver(f.operator, safetytypes.NewMsgTrip(f.operator.String(), safetytypes.Wildcard))
	require.True(t, f.isAllowed(execURL))
	require.False(t, f.isAllowed(msgSendURL))
	f.requireRejected(f.other, safetytypes.ErrMsgDisabled, exec(send(f.user, f.other)))

	bz := f.encode(f.other, exec(resetAll))
	cres := f.checkTx(bz)
	require.Zero(t, cres.Code, cres.Log)
	res := f.finalize(bz).TxResults[0]
	require.Zero(t, res.Code, res.Log)
	require.Empty(t, f.disabled())

	// the operator pins MsgExec beside "*"; the authority loses its path
	f.mustDeliver(f.operator, safetytypes.NewMsgTrip(f.operator.String(), safetytypes.Wildcard, execURL))
	require.False(t, f.isAllowed(execURL))
	f.requireRejected(f.other, safetytypes.ErrMsgDisabled, exec(resetAll))
	require.ElementsMatch(t, []string{safetytypes.Wildcard, execURL}, f.disabled())

	f.mustDeliver(f.operator, safetytypes.NewMsgReset(f.operator.String(), execURL))
	f.mustDeliver(f.other, exec(resetAll))
	require.Empty(t, f.disabled())
}
