package safety_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	safetytypes "github.com/cosmos/cosmos-sdk/x/safety/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	msgWithdrawRewardURL = "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward"
	msgVoteURL           = "/cosmos.gov.v1.MsgVote"
	// hostMsgURL is the test msg whose msg server calls the SDK msg servers
	// directly, as MsgEthereumTx reaches EVM precompiles.
	hostMsgURL = "/testpb.MsgCreateDog"
)

// directURLs lists the msg servers the precompile calls, in call order.
var directURLs = []string{msgSendURL, msgDelegateURL, msgWithdrawRewardURL, msgVoteURL}

// site is a place in FinalizeBlock where the precompile can run.
type site string

const (
	// siteHostMsg is the msg server of the host msg, dispatched by the
	// MsgServiceRouter inside a tx: the path an EVM precompile takes.
	siteHostMsg site = "host msg server"
	// sitePostHandler is the post handler of a tx, on the Context the tx
	// msgs ran on.
	sitePostHandler site = "post handler"
	// sitePreBlocker is the PreBlocker, on the block Context.
	sitePreBlocker site = "pre blocker"
	// siteBeginBlocker is the BeginBlocker, on the block Context.
	siteBeginBlocker site = "begin blocker"
	// siteEndBlocker is the EndBlocker, on the block Context.
	siteEndBlocker site = "end blocker"
)

// precompile stands in for a cosmos/evm precompile. Armed for a site, it
// calls the bank, staking, distribution and gov msg servers as Go methods
// on the Context it receives there, without the MsgServiceRouter, records
// each call's error, and returns success, so the state each call writes
// commits with the tx or block.
type precompile struct {
	f          *fixture
	proposalID uint64

	armed site
	// runs counts the runs so far; each run votes with its own metadata.
	runs int
	errs map[string]error
	// breaker and checkedURL record the circuit breaker and the router mark
	// of the Context the last run received.
	breaker    bool
	checkedURL string
}

func (p *precompile) maybeRun(ctx sdk.Context, at site) {
	if p.armed != at {
		return
	}
	p.armed = ""
	p.runs++
	p.breaker = ctx.CircuitBreaker() != nil
	p.checkedURL = ctx.CheckedMsgTypeURL()

	f := p.f
	val := f.validator()
	p.errs = map[string]error{}
	_, p.errs[msgSendURL] = bankkeeper.NewMsgServerImpl(f.app.BankKeeper).Send(ctx, send(f.user, f.other))
	_, p.errs[msgDelegateURL] = stakingkeeper.NewMsgServerImpl(f.app.StakingKeeper).Delegate(ctx, stakingDelegate(f.user, val.String()))
	_, p.errs[msgWithdrawRewardURL] = distrkeeper.NewMsgServerImpl(f.app.DistrKeeper).WithdrawDelegatorReward(ctx,
		distrtypes.NewMsgWithdrawDelegatorReward(f.voter.String(), val.String()))
	_, p.errs[msgVoteURL] = govkeeper.NewMsgServerImpl(&f.app.GovKeeper).Vote(ctx,
		govv1.NewMsgVote(f.user.addr, p.proposalID, govv1.OptionAbstain, p.voteMetadata()))
}

// voteMetadata is the metadata of the direct vote of the current run.
func (p *precompile) voteMetadata() string { return fmt.Sprintf("run %d", p.runs) }

// hostMsgServer is the msg server of the host msg.
type hostMsgServer struct{ p *precompile }

func (s hostMsgServer) CreateDog(ctx context.Context, msg *testdata.MsgCreateDog) (*testdata.MsgCreateDogResponse, error) {
	s.p.maybeRun(sdk.UnwrapSDKContext(ctx), siteHostMsg)
	return &testdata.MsgCreateDogResponse{Name: msg.Dog.Name}, nil
}

// installPrecompile registers the host msg service and wraps the post
// handler and block hooks of f.app, so the precompile runs inside
// FinalizeBlock. SimApp's post handler chain is empty.
func installPrecompile(f *fixture) *precompile {
	p := &precompile{f: f}
	testdata.RegisterInterfaces(f.app.InterfaceRegistry())
	testdata.RegisterMsgServer(f.app.MsgServiceRouter(), hostMsgServer{p})
	f.app.SetPostHandler(func(ctx sdk.Context, _ sdk.Tx, _, _ bool) (sdk.Context, error) {
		if ctx.ExecMode() == sdk.ExecModeFinalize {
			p.maybeRun(ctx, sitePostHandler)
		}
		return ctx, nil
	})
	f.app.SetPreBlocker(func(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
		p.maybeRun(ctx, sitePreBlocker)
		return f.app.PreBlocker(ctx, req)
	})
	f.app.SetBeginBlocker(func(ctx sdk.Context) (sdk.BeginBlock, error) {
		p.maybeRun(ctx, siteBeginBlocker)
		return f.app.BeginBlocker(ctx)
	})
	f.app.SetEndBlocker(func(ctx sdk.Context) (sdk.EndBlock, error) {
		p.maybeRun(ctx, siteEndBlocker)
		return f.app.EndBlocker(ctx)
	})
	return p
}

// newPrecompileFixture starts a SimApp chain with the precompile installed,
// appoints f.operator with a "*" scope, and opens a text proposal for the
// direct votes. withoutBreaker leaves BaseApp without a circuit breaker, as
// in an app that does not wire x/safety.
func newPrecompileFixture(t *testing.T, withoutBreaker bool) (*fixture, *precompile) {
	t.Helper()
	var p *precompile
	f := newFixtureWithApp(t, nil, nil, func(f *fixture) {
		p = installPrecompile(f)
		if withoutBreaker {
			f.app.SetCircuitBreaker(nil)
		}
	})
	f.setParams(f.operator.String(), safetytypes.Wildcard)
	p.openProposal()
	return f, p
}

// openProposal opens a text proposal for the direct votes, in place of one
// whose voting period has ended.
func (p *precompile) openProposal() {
	f := p.f
	f.t.Helper()
	govParams, err := f.app.GovKeeper.Params.Get(f.ctx())
	require.NoError(f.t, err)
	p.proposalID, err = f.app.GovKeeper.ProposalID.Peek(f.ctx())
	require.NoError(f.t, err)
	submit, err := govv1.NewMsgSubmitProposal(nil, govParams.MinDeposit, f.voter.String(), "direct", "direct", "direct", false)
	require.NoError(f.t, err)
	f.mustDeliver(f.voter, submit)
}

func (f *fixture) validator() sdk.ValAddress {
	return sdk.ValAddress(f.valSet.Validators[0].Address)
}

// directState is the state each direct call writes.
type directState struct {
	// otherBalance is what MsgSend from the user pays into.
	otherBalance sdkmath.Int
	// userShares is the user's delegation to the validator.
	userShares sdkmath.LegacyDec
	// voterPeriod is the reward period the voter's delegation starts from,
	// which a reward withdrawal advances.
	voterPeriod uint64
	// userVote is the metadata of the user's vote on the proposal.
	userVote string
}

func (f *fixture) directState(proposalID uint64) directState {
	f.t.Helper()
	ctx := f.ctx()
	s := directState{otherBalance: f.balance(f.other.addr), userShares: sdkmath.LegacyZeroDec()}

	del, err := f.app.StakingKeeper.GetDelegation(ctx, f.user.addr, f.validator())
	if !errors.Is(err, stakingtypes.ErrNoDelegation) {
		require.NoError(f.t, err)
		s.userShares = del.Shares
	}

	info, err := f.app.DistrKeeper.GetDelegatorStartingInfo(ctx, f.validator(), f.voter.addr)
	require.NoError(f.t, err)
	s.voterPeriod = info.PreviousPeriod

	vote, err := f.app.GovKeeper.Votes.Get(ctx, collections.Join(proposalID, f.user.addr))
	if !errors.Is(err, collections.ErrNotFound) {
		require.NoError(f.t, err)
		s.userVote = vote.Metadata
	}
	return s
}

// runAt arms the precompile for at, runs the block that reaches it, and
// requires each direct call in blocked to fail with ErrMsgDisabled and leave
// its state as it was, and every other direct call to succeed and write its
// state. It returns the precompile for further checks on the run.
func (p *precompile) runAt(t *testing.T, at site, blocked ...string) *precompile {
	t.Helper()
	f := p.f
	before := f.directState(p.proposalID)

	p.armed = at
	switch at {
	case siteHostMsg:
		f.mustDeliver(f.voter, &testdata.MsgCreateDog{Dog: &testdata.Dog{Name: "spot"}, Owner: f.voter.String()})
	case sitePostHandler:
		// a gov vote is protected, so this tx runs under every trip
		f.mustDeliver(f.voter, govv1.NewMsgVote(f.voter.addr, p.proposalID, govv1.OptionYes, ""))
	case sitePreBlocker, siteBeginBlocker, siteEndBlocker:
		f.finalize()
	}
	require.Empty(t, p.armed, "the precompile did not run at the %s", at)

	after := f.directState(p.proposalID)
	for _, url := range directURLs {
		err := p.errs[url]
		isBlocked := false
		for _, b := range blocked {
			isBlocked = isBlocked || b == url
		}
		if isBlocked {
			require.ErrorIs(t, err, safetytypes.ErrMsgDisabled, url)
			require.ErrorContains(t, err, url)
		} else {
			require.NoError(t, err, url)
		}

		switch url {
		case msgSendURL:
			if isBlocked {
				require.Equal(t, before.otherBalance, after.otherBalance, "a blocked send moved funds")
			} else {
				require.Equal(t, before.otherBalance.AddRaw(1), after.otherBalance)
			}
		case msgDelegateURL:
			if isBlocked {
				require.Equal(t, before.userShares, after.userShares, "a blocked delegation changed shares")
			} else {
				require.True(t, after.userShares.GT(before.userShares), "the delegation added no shares")
			}
		case msgWithdrawRewardURL:
			if isBlocked {
				require.Equal(t, before.voterPeriod, after.voterPeriod, "a blocked withdrawal advanced the period")
			} else {
				require.Greater(t, after.voterPeriod, before.voterPeriod)
			}
		case msgVoteURL:
			if isBlocked {
				require.Equal(t, before.userVote, after.userVote, "a blocked vote was stored")
			} else {
				require.Equal(t, p.voteMetadata(), after.userVote)
			}
		}
	}
	return p
}

// requireAllRun runs the precompile at each of sites and requires every direct
// call to succeed.
func (p *precompile) requireAllRun(t *testing.T, sites ...site) {
	t.Helper()
	for _, at := range sites {
		p.runAt(t, at)
	}
}

var (
	blockHookSites = []site{sitePreBlocker, siteBeginBlocker, siteEndBlocker}
	allSites       = append([]site{siteHostMsg, sitePostHandler}, blockHookSites...)
)

// TestDirectCallsSelectiveTrip shows that a trip of MsgSend stops a direct
// call of the bank msg server made on a tx or block Context inside
// FinalizeBlock, including from the msg server of a routed msg, whose router
// mark names the outer msg only, while the direct calls the trip does not
// cover still run, and that all of them run again after the reset.
func TestDirectCallsSelectiveTrip(t *testing.T) {
	f, p := newPrecompileFixture(t, false)
	testDirectCallsSelectiveTrip(t, f, p)
}

// TestDirectCallsSelectiveTripAfterProcessProposal is
// TestDirectCallsSelectiveTrip with every block run the way CometBFT runs it
// on a validator: ProcessProposal resets the block state, and FinalizeBlock
// reuses it.
func TestDirectCallsSelectiveTripAfterProcessProposal(t *testing.T) {
	f, p := newPrecompileFixture(t, false)
	f.processFirst = true
	testDirectCallsSelectiveTrip(t, f, p)
}

func testDirectCallsSelectiveTrip(t *testing.T, f *fixture, p *precompile) {
	t.Helper()

	for _, at := range allSites {
		p.runAt(t, at)
		require.True(t, p.breaker, "BaseApp must attach the breaker at the %s", at)
	}
	require.Equal(t, hostMsgURL, p.runAt(t, siteHostMsg).checkedURL)

	f.mustDeliver(f.operator, safetytypes.NewMsgTrip(f.operator.String(), msgSendURL))
	for _, at := range allSites {
		t.Run(string(at), func(t *testing.T) {
			f.t = t
			p.runAt(t, at, msgSendURL)
		})
	}
	f.t = t

	f.mustDeliver(f.operator, safetytypes.NewMsgReset(f.operator.String(), msgSendURL))
	p.requireAllRun(t, allSites...)
}

// TestDirectCallsWildcardTrip shows that under "*" the host msg itself is
// rejected, so a precompile behind it never runs, and that a direct call made
// on a tx or block Context that still runs, behind a protected msg or in the
// EndBlocker, fails for every msg server but the protected gov one. After the
// reset every direct call runs again.
func TestDirectCallsWildcardTrip(t *testing.T) {
	f, p := newPrecompileFixture(t, false)
	f.mustDeliver(f.operator, safetytypes.NewMsgTrip(f.operator.String(), safetytypes.Wildcard))

	p.armed = siteHostMsg
	f.requireRejected(f.voter, safetytypes.ErrMsgDisabled, &testdata.MsgCreateDog{Dog: &testdata.Dog{Name: "spot"}, Owner: f.voter.String()})
	require.Equal(t, siteHostMsg, p.armed, "the host msg ran under *")
	p.armed = ""

	for _, at := range append([]site{sitePostHandler}, blockHookSites...) {
		t.Run(string(at), func(t *testing.T) {
			f.t = t
			p.runAt(t, at, msgSendURL, msgDelegateURL, msgWithdrawRewardURL)
		})
	}
	f.t = t

	f.mustDeliver(f.operator, safetytypes.NewMsgReset(f.operator.String(), safetytypes.Wildcard))
	p.requireAllRun(t, allSites...)
}

// TestDirectCallsWithoutBreaker shows that the msg server guard allows every
// direct call when BaseApp has no circuit breaker, as in an app that does not
// wire x/safety, even with x/safety state that disables MsgSend and "*".
func TestDirectCallsWithoutBreaker(t *testing.T) {
	f, p := newPrecompileFixture(t, true)

	f.mustDeliver(f.operator, safetytypes.NewMsgTrip(f.operator.String(), msgSendURL))
	p.requireAllRun(t, allSites...)
	require.False(t, p.breaker)

	f.mustDeliver(f.operator, safetytypes.NewMsgTrip(f.operator.String(), safetytypes.Wildcard))
	p.requireAllRun(t, append([]site{sitePostHandler}, blockHookSites...)...)
	require.False(t, p.breaker)
}
