package safety_test

import (
	"context"
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/simapp"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	safetytypes "github.com/cosmos/cosmos-sdk/x/safety/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	msgSendURL      = "/cosmos.bank.v1beta1.MsgSend"
	msgMultiSendURL = "/cosmos.bank.v1beta1.MsgMultiSend"
	msgDelegateURL  = "/cosmos.staking.v1beta1.MsgDelegate"
	distrParamsURL  = "/cosmos.distribution.v1beta1.MsgUpdateParams"
	authzGrantURL   = "/cosmos.authz.v1beta1.MsgGrant"
	authzExecURL    = "/cosmos.authz.v1beta1.MsgExec"
	authzRevokeURL  = "/cosmos.authz.v1beta1.MsgRevoke"
)

var (
	govAddr = authtypes.NewModuleAddress(govtypes.ModuleName).String()
	txFee   = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000))
)

// account is a funded genesis account with its signing key.
type account struct {
	priv *secp256k1.PrivKey
	addr sdk.AccAddress
}

func (a account) String() string { return a.addr.String() }

// fixture is a full SimApp chain driven through ABCI: txs are signed,
// encoded, and run through CheckTx and FinalizeBlock, and gov proposals go
// through submission, voting, and the gov EndBlocker.
type fixture struct {
	t      *testing.T
	app    *simapp.SimApp
	db     dbm.DB
	home   string
	valSet *cmttypes.ValidatorSet
	now    time.Time
	// initialHeight is the first block height of a chain started from an
	// export, zero otherwise.
	initialHeight int64

	// voter is the only delegator of the only validator, so its vote alone
	// passes a proposal.
	voter account
	// user, operator, and other are plain funded accounts.
	user     account
	operator account
	other    account
}

// defaultConsensusParams returns a copy of the consensus params SimApp tests
// start from. They set the consensus AuthorityParams override to the gov
// module account.
func defaultConsensusParams() *cmtproto.ConsensusParams {
	cp := *simtestutil.DefaultConsensusParams
	auth := *cp.Authority
	cp.Authority = &auth
	return &cp
}

// newFixture starts a SimApp chain with the given consensus params, or the
// SimApp test defaults when cp is nil. mutate may change the genesis state
// before InitChain.
func newFixture(t *testing.T, cp *cmtproto.ConsensusParams, mutate func(simapp.GenesisState)) *fixture {
	t.Helper()
	return newFixtureWithSetup(t, cp, mutate, nil)
}

// newFixtureWithSetup is newFixture with setup run on the SimApp before it
// loads and seals, so a test can wrap its block hooks or ante handler.
func newFixtureWithSetup(t *testing.T, cp *cmtproto.ConsensusParams, mutate func(simapp.GenesisState), setup func(*simapp.SimApp)) *fixture {
	t.Helper()
	if cp == nil {
		cp = defaultConsensusParams()
	}

	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)
	valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{cmttypes.NewValidator(pubKey, 1)})

	newAccount := func() account {
		priv := secp256k1.GenPrivKey()
		return account{priv: priv, addr: sdk.AccAddress(priv.PubKey().Address())}
	}
	f := &fixture{
		t:        t,
		valSet:   valSet,
		now:      time.Unix(1_700_000_000, 0).UTC(),
		voter:    newAccount(),
		user:     newAccount(),
		operator: newAccount(),
		other:    newAccount(),
	}

	funds := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100_000_000_000_000)))
	var genAccs []authtypes.GenesisAccount
	var balances []banktypes.Balance
	// the voter must come first: GenesisStateWithValSet delegates from
	// genAccs[0]
	for _, a := range []account{f.voter, f.user, f.operator, f.other} {
		genAccs = append(genAccs, authtypes.NewBaseAccount(a.addr, a.priv.PubKey(), 0, 0))
		balances = append(balances, banktypes.Balance{Address: a.String(), Coins: funds})
	}

	f.db, f.home = dbm.NewMemDB(), t.TempDir()
	f.app = simapp.NewSimApp(log.NewNopLogger(), f.db, setup == nil, simtestutil.NewAppOptionsWithFlagHome(f.home))
	if setup != nil {
		setup(f.app)
		require.NoError(t, f.app.LoadLatestVersion())
	}
	genesis := f.app.DefaultGenesis()
	genesis, err = simtestutil.GenesisStateWithValSet(f.app.AppCodec(), genesis, valSet, genAccs, balances...)
	require.NoError(t, err)
	if mutate != nil {
		mutate(genesis)
	}
	stateBytes, err := json.MarshalIndent(genesis, "", " ")
	require.NoError(t, err)

	_, err = f.app.InitChain(&abci.RequestInitChain{
		Time:            f.now,
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: cp,
		AppStateBytes:   stateBytes,
	})
	require.NoError(t, err)
	f.finalize()
	return f
}

// finalize runs one block with the given txs at the next block time and
// commits it.
func (f *fixture) finalize(txs ...[]byte) *abci.ResponseFinalizeBlock {
	f.t.Helper()
	f.now = f.now.Add(5 * time.Second)
	return f.finalizeAt(f.now, txs...)
}

func (f *fixture) finalizeAt(at time.Time, txs ...[]byte) *abci.ResponseFinalizeBlock {
	f.t.Helper()
	f.now = at
	height := f.app.LastBlockHeight() + 1
	if f.initialHeight > height {
		height = f.initialHeight
	}
	res, err := f.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             height,
		Time:               at,
		Hash:               f.app.LastCommitID().Hash,
		NextValidatorsHash: f.valSet.Hash(),
		Txs:                txs,
	})
	require.NoError(f.t, err)
	require.Len(f.t, res.TxResults, len(txs))
	_, err = f.app.Commit()
	require.NoError(f.t, err)
	return res
}

// ctx returns a read-only context on the last committed state.
func (f *fixture) ctx() sdk.Context {
	return sdk.NewContext(
		f.app.CommitMultiStore().CacheMultiStore(),
		cmtproto.Header{Height: f.app.LastBlockHeight(), Time: f.now},
		false, log.NewNopLogger(),
	)
}

// checkCtx returns the CheckTx state, which reflects every tx CheckTx admitted
// since the last commit, including the fees their ante handlers deducted.
func (f *fixture) checkCtx() sdk.Context { return f.app.NewContext(true) }

// seq returns the account's committed sequence.
func (f *fixture) seq(a account) uint64 {
	acc := f.app.AccountKeeper.GetAccount(f.ctx(), a.addr)
	require.NotNil(f.t, acc)
	return acc.GetSequence()
}

// encode signs msgs as a tx from signer at its committed account number and
// sequence.
func (f *fixture) encode(signer account, msgs ...sdk.Msg) []byte {
	f.t.Helper()
	return f.encodeAt(signer, f.seq(signer), msgs...)
}

// encodeAt signs msgs as a tx from signer at sequence seq.
func (f *fixture) encodeAt(signer account, seq uint64, msgs ...sdk.Msg) []byte {
	f.t.Helper()
	acc := f.app.AccountKeeper.GetAccount(f.ctx(), signer.addr)
	require.NotNil(f.t, acc)
	txConfig := f.app.TxConfig()
	tx, err := simtestutil.GenSignedMockTx(rand.New(rand.NewSource(1)), txConfig, msgs, txFee,
		simtestutil.DefaultGenTxGas, "", []uint64{acc.GetAccountNumber()}, []uint64{seq}, signer.priv)
	require.NoError(f.t, err)
	bz, err := txConfig.TxEncoder()(tx)
	require.NoError(f.t, err)
	return bz
}

func (f *fixture) checkTx(bz []byte) *abci.ResponseCheckTx {
	f.t.Helper()
	res, err := f.app.CheckTx(&abci.RequestCheckTx{Tx: bz, Type: abci.CheckTxType_New})
	require.NoError(f.t, err)
	return res
}

// deliver signs msgs from signer and runs them in their own block.
func (f *fixture) deliver(signer account, msgs ...sdk.Msg) *abci.ExecTxResult {
	f.t.Helper()
	return f.finalize(f.encode(signer, msgs...)).TxResults[0]
}

// deliverTxs runs txs in one block and returns their results.
func (f *fixture) deliverTxs(txs ...[]byte) []*abci.ExecTxResult {
	f.t.Helper()
	return f.finalize(txs...).TxResults
}

// mustDeliver delivers msgs and requires success.
func (f *fixture) mustDeliver(signer account, msgs ...sdk.Msg) *abci.ExecTxResult {
	f.t.Helper()
	res := f.deliver(signer, msgs...)
	require.Zero(f.t, res.Code, res.Log)
	return res
}

// requireRejected requires that a tx with msgs from signer fails CheckTx with
// wantErr before any fee is taken, and fails FinalizeBlock with wantErr at the
// router, where it pays its fee and uses its sequence like any failed tx, so
// the same bytes cannot run again.
func (f *fixture) requireRejected(signer account, wantErr interface {
	ABCICode() uint32
	Codespace() string
}, msgs ...sdk.Msg,
) {
	f.t.Helper()
	bz := f.requireCheckTxRejected(signer, wantErr, msgs...)
	before := f.balance(signer.addr)
	seq := f.app.AccountKeeper.GetAccount(f.ctx(), signer.addr).GetSequence()

	dres := f.finalize(bz).TxResults[0]
	require.Equal(f.t, wantErr.ABCICode(), dres.Code, dres.Log)
	require.Equal(f.t, wantErr.Codespace(), dres.Codespace, dres.Log)
	require.Equal(f.t, before.Sub(txFee.AmountOf(sdk.DefaultBondDenom)), f.balance(signer.addr), "the fee must be taken")
	require.Equal(f.t, seq+1, f.app.AccountKeeper.GetAccount(f.ctx(), signer.addr).GetSequence(), "the sequence must be used")

	replay := f.finalize(bz).TxResults[0]
	require.Equal(f.t, sdkerrors.ErrWrongSequence.ABCICode(), replay.Code, replay.Log)
}

// requireCheckTxRejected requires that a tx with msgs from signer fails CheckTx
// with wantErr and that no fee is taken, and returns the tx bytes.
func (f *fixture) requireCheckTxRejected(signer account, wantErr interface {
	ABCICode() uint32
	Codespace() string
}, msgs ...sdk.Msg,
) []byte {
	f.t.Helper()
	before := f.balance(signer.addr)
	bz := f.encode(signer, msgs...)

	cres := f.checkTx(bz)
	require.Equal(f.t, wantErr.ABCICode(), cres.Code, cres.Log)
	require.Equal(f.t, wantErr.Codespace(), cres.Codespace, cres.Log)
	require.Equal(f.t, before, f.balance(signer.addr), "no fee may be taken")
	return bz
}

// requireFails delivers msgs and requires the tx to fail with wantErr.
func (f *fixture) requireFails(signer account, wantErr interface {
	ABCICode() uint32
	Codespace() string
}, msgs ...sdk.Msg,
) *abci.ExecTxResult {
	f.t.Helper()
	res := f.deliver(signer, msgs...)
	require.Equal(f.t, wantErr.ABCICode(), res.Code, res.Log)
	require.Equal(f.t, wantErr.Codespace(), res.Codespace, res.Log)
	return res
}

func (f *fixture) balance(addr sdk.AccAddress) sdkmath.Int {
	return f.app.BankKeeper.GetBalance(f.ctx(), addr, sdk.DefaultBondDenom).Amount
}

func (f *fixture) params() safetytypes.Params {
	f.t.Helper()
	p, err := f.app.SafetyKeeper.Params.Get(f.ctx())
	require.NoError(f.t, err)
	return p
}

// disabled returns the disabled list through the gRPC query, in store order.
func (f *fixture) disabled() []string {
	f.t.Helper()
	var resp safetytypes.QueryDisabledListResponse
	f.query("/cosmos.safety.v1.Query/DisabledList", &safetytypes.QueryDisabledListRequest{}, &resp)
	return resp.DisabledPatterns
}

// isAllowed asks the IsAllowed gRPC query.
func (f *fixture) isAllowed(typeURL string) bool {
	f.t.Helper()
	var resp safetytypes.QueryIsAllowedResponse
	f.query("/cosmos.safety.v1.Query/IsAllowed", &safetytypes.QueryIsAllowedRequest{MsgTypeUrl: typeURL}, &resp)
	return resp.Allowed
}

// query runs an ABCI gRPC query against the latest committed state.
func (f *fixture) query(path string, req, resp proto.Message) {
	f.t.Helper()
	bz, err := proto.Marshal(req)
	require.NoError(f.t, err)
	res, err := f.app.Query(context.Background(), &abci.RequestQuery{Path: path, Data: bz})
	require.NoError(f.t, err)
	require.Zero(f.t, res.Code, res.Log)
	require.NoError(f.t, proto.Unmarshal(res.Value, resp))
}

// submitProposal submits msgs as a gov proposal from the voter with the full
// deposit and votes yes, and returns its id. The proposal stays in voting
// until endVoting.
func (f *fixture) submitProposal(msgs ...sdk.Msg) uint64 {
	f.t.Helper()
	govParams, err := f.app.GovKeeper.Params.Get(f.ctx())
	require.NoError(f.t, err)
	id, err := f.app.GovKeeper.ProposalID.Peek(f.ctx())
	require.NoError(f.t, err)

	submit, err := govv1.NewMsgSubmitProposal(msgs, govParams.MinDeposit, f.voter.String(), "", "safety", "safety", false)
	require.NoError(f.t, err)
	f.mustDeliver(f.voter, submit)
	f.mustDeliver(f.voter, govv1.NewMsgVote(f.voter.addr, id, govv1.OptionYes, ""))
	f.requireProposalStatus(id, govv1.StatusVotingPeriod)
	return id
}

// endVoting runs an empty block past the voting period, so the gov
// EndBlocker tallies and executes every proposal in voting.
func (f *fixture) endVoting() *abci.ResponseFinalizeBlock {
	f.t.Helper()
	govParams, err := f.app.GovKeeper.Params.Get(f.ctx())
	require.NoError(f.t, err)
	return f.finalizeAt(f.now.Add(*govParams.VotingPeriod + time.Second))
}

func (f *fixture) proposal(id uint64) govv1.Proposal {
	f.t.Helper()
	p, err := f.app.GovKeeper.Proposals.Get(f.ctx(), id)
	require.NoError(f.t, err)
	return p
}

func (f *fixture) requireProposalStatus(id uint64, status govv1.ProposalStatus) govv1.Proposal {
	f.t.Helper()
	p := f.proposal(id)
	require.Equal(f.t, status, p.Status, "proposal %d failed reason: %q", id, p.FailedReason)
	return p
}

// propose runs msgs as a gov proposal through voting and returns the
// proposal after the gov EndBlocker has tallied and executed it, along with
// that block's response.
func (f *fixture) propose(msgs ...sdk.Msg) (govv1.Proposal, *abci.ResponseFinalizeBlock) {
	f.t.Helper()
	id := f.submitProposal(msgs...)
	res := f.endVoting()
	return f.proposal(id), res
}

// mustPass runs a proposal and requires it to pass and execute.
func (f *fixture) mustPass(msgs ...sdk.Msg) *abci.ResponseFinalizeBlock {
	f.t.Helper()
	proposal, res := f.propose(msgs...)
	require.Equal(f.t, govv1.StatusPassed, proposal.Status, proposal.FailedReason)
	return res
}

// mustFailExecution runs a proposal and requires it to pass the vote but fail
// when the gov module executes its msgs. It returns the failure reason.
func (f *fixture) mustFailExecution(msgs ...sdk.Msg) string {
	f.t.Helper()
	proposal, _ := f.propose(msgs...)
	require.Equal(f.t, govv1.StatusFailed, proposal.Status)
	return proposal.FailedReason
}

// setParams changes the x/safety params through a gov proposal.
func (f *fixture) setParams(operator string, scope ...string) {
	f.t.Helper()
	p := f.params()
	p.Operator = operator
	p.Scope = scope
	f.mustPass(&safetytypes.MsgUpdateParams{Authority: govAddr, Params: p})
	require.Equal(f.t, p, f.params())
}

// accountOf returns the fixture account with address addr.
func (f *fixture) accountOf(addr string) account {
	f.t.Helper()
	for _, a := range []account{f.voter, f.user, f.operator, f.other} {
		if a.String() == addr {
			return a
		}
	}
	require.FailNow(f.t, "no fixture account", addr)
	return account{}
}

// trip trips patterns as the current operator and requires success.
func (f *fixture) trip(patterns ...string) {
	f.t.Helper()
	op := f.accountOf(f.params().Operator)
	f.mustDeliver(op, safetytypes.NewMsgTrip(op.String(), patterns...))
}

// reset resets patterns as the current operator and requires success.
func (f *fixture) reset(patterns ...string) {
	f.t.Helper()
	op := f.accountOf(f.params().Operator)
	f.mustDeliver(op, safetytypes.NewMsgReset(op.String(), patterns...))
}

// send is a 1stake bank send from one account to another.
func send(from, to account) *banktypes.MsgSend {
	return banktypes.NewMsgSend(from.addr, to.addr, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)))
}

// typedEvents returns the events of type T among events.
func typedEvents[T proto.Message](t *testing.T, events []abci.Event) []T {
	t.Helper()
	var out []T
	var zero T
	name := proto.MessageName(zero)
	for _, e := range events {
		if e.Type != name {
			continue
		}
		// strip the attributes baseapp appends, such as mode and msg_index,
		// which are not fields of the typed event
		attrs := make([]abci.EventAttribute, 0, len(e.Attributes))
		for _, a := range e.Attributes {
			if a.Key != "mode" && a.Key != "msg_index" {
				attrs = append(attrs, a)
			}
		}
		msg, err := sdk.ParseTypedEvent(abci.Event{Type: e.Type, Attributes: attrs})
		require.NoError(t, err)
		out = append(out, msg.(T))
	}
	return out
}

// multiSend is a 1stake bank multi-send from one account to another.
func multiSend(from, to account) *banktypes.MsgMultiSend {
	coins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1))
	return banktypes.NewMsgMultiSend(banktypes.NewInput(from.addr, coins), []banktypes.Output{banktypes.NewOutput(to.addr, coins)})
}

// sendFromGov is a 1stake bank send from the gov module account, as a
// proposal would contain it.
func sendFromGov(to account) *banktypes.MsgSend {
	return &banktypes.MsgSend{
		FromAddress: govAddr,
		ToAddress:   to.String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)),
	}
}

// stakingDelegate is a 1stake delegation from an account to a validator.
func stakingDelegate(from account, valAddr string) *stakingtypes.MsgDelegate {
	return stakingtypes.NewMsgDelegate(from.String(), valAddr, sdk.NewInt64Coin(sdk.DefaultBondDenom, 1))
}
