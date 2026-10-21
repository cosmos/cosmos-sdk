package safety_test

import (
	"math/rand"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	txconfigv1 "cosmossdk.io/api/cosmos/tx/config/v1"
	"cosmossdk.io/core/appconfig"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log/v2"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/simapp"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	groupmodulev1 "github.com/cosmos/cosmos-sdk/enterprise/group/api/cosmos/group/module/v1"
	"github.com/cosmos/cosmos-sdk/enterprise/group/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/enterprise/group/x/group/keeper"
	_ "github.com/cosmos/cosmos-sdk/enterprise/group/x/group/module" // import as blank for app wiring
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	safetykeeper "github.com/cosmos/cosmos-sdk/x/safety/keeper"
	safetytestutil "github.com/cosmos/cosmos-sdk/x/safety/testutil"
	safetytypes "github.com/cosmos/cosmos-sdk/x/safety/types"
)

// groupFixture is a depinject app with auth, bank, staking, consensus, group,
// authz and safety, and the SimApp ante chain, which runs SafetyDecorator
// before fee deduction. SimApp does not wire enterprise/group, so the group
// paths run here. The chain starts without the consensus AuthorityParams
// override, so the x/safety params authority is the authority.
type groupFixture struct {
	t        *testing.T
	app      *runtime.App
	txConfig client.TxConfig

	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.Keeper
	groupKeeper   groupkeeper.Keeper
	safetyKeeper  safetykeeper.Keeper

	// member is the only member of the group and its admin until the policy
	// takes over; it is also the x/safety authority at genesis.
	member   account
	operator account
	// policy is the group policy account, admin of its own group, which
	// becomes the x/safety authority.
	policy  sdk.AccAddress
	groupID uint64
	// proposals counts the group proposals submitted so far; ids start at 1.
	proposals uint64
	now       time.Time
}

func groupModule() configurator.ModuleOption {
	return func(config *configurator.Config) {
		config.ModuleConfigs["group"] = &appv1alpha1.ModuleConfig{
			Name:   "group",
			Config: appconfig.WrapAny(&groupmodulev1.Module{}),
		}
	}
}

// txModuleSkipAnte wires x/auth/tx without its default ante handler, so the
// fixture installs the SimApp chain.
func txModuleSkipAnte() configurator.ModuleOption {
	return func(config *configurator.Config) {
		config.ModuleConfigs["tx"] = &appv1alpha1.ModuleConfig{
			Name:   "tx",
			Config: appconfig.WrapAny(&txconfigv1.Config{SkipAnteHandler: true}),
		}
	}
}

func newGroupFixture(t *testing.T) *groupFixture {
	t.Helper()
	newAccount := func() account {
		priv := secp256k1.GenPrivKey()
		return account{priv: priv, addr: sdk.AccAddress(priv.PubKey().Address())}
	}
	f := &groupFixture{t: t, member: newAccount(), operator: newAccount(), now: time.Unix(1_700_000_000, 0).UTC()}

	var (
		appBuilder *runtime.AppBuilder
		cdc        codec.Codec
	)
	require.NoError(t, depinject.Inject(
		depinject.Configs(
			configurator.NewAppConfig(
				configurator.AuthModule(),
				configurator.BankModule(),
				configurator.StakingModule(),
				configurator.ConsensusModule(),
				txModuleSkipAnte(),
				groupModule(),
				configurator.AuthzModule(),
				safetytestutil.SafetyModule(f.member.String()),
			),
			depinject.Supply(log.NewNopLogger()),
		),
		&appBuilder, &cdc, &f.txConfig, &f.accountKeeper, &f.bankKeeper, &f.groupKeeper, &f.safetyKeeper,
	))

	anteHandler, err := simapp.NewAnteHandler(simapp.HandlerOptions{
		HandlerOptions: authante.HandlerOptions{
			AccountKeeper:   f.accountKeeper,
			BankKeeper:      f.bankKeeper,
			SignModeHandler: f.txConfig.SignModeHandler(),
			SigGasConsumer:  authante.DefaultSigVerificationGasConsumer,
		},
		SafetyKeeper: f.safetyKeeper,
	})
	require.NoError(t, err)
	f.app = appBuilder.Build(dbm.NewMemDB(), func(app *baseapp.BaseApp) {
		app.SetAnteHandler(anteHandler)
	})
	require.NoError(t, f.app.Load(true))

	valSet, err := simtestutil.CreateRandomValidatorSet()
	require.NoError(t, err)
	funds := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(1_000_000_000_000)))
	var (
		genAccs  []authtypes.GenesisAccount
		balances []banktypes.Balance
	)
	for i, a := range []account{f.member, f.operator} {
		genAccs = append(genAccs, authtypes.NewBaseAccount(a.addr, nil, uint64(i), 0))
		balances = append(balances, banktypes.Balance{Address: a.String(), Coins: funds})
	}
	genesis, err := simtestutil.GenesisStateWithValSet(cdc, f.app.DefaultGenesis(), valSet, genAccs, balances...)
	require.NoError(t, err)
	stateBytes, err := cmtjson.MarshalIndent(genesis, "", " ")
	require.NoError(t, err)

	cp := *simtestutil.DefaultConsensusParams
	cp.Authority = nil
	_, err = f.app.InitChain(&abci.RequestInitChain{
		Time:            f.now,
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: &cp,
		AppStateBytes:   stateBytes,
	})
	require.NoError(t, err)
	f.finalize()

	// the member creates a group whose policy administers the group and
	// itself, and one vote passes a proposal
	create, err := group.NewMsgCreateGroupWithPolicy(f.member.String(),
		[]group.MemberRequest{{Address: f.member.String(), Weight: "1"}}, "", "", true,
		group.NewThresholdDecisionPolicy("1", time.Hour, 0))
	require.NoError(t, err)
	f.requireTx(f.member, "", create)
	res, err := f.groupKeeper.GroupPoliciesByGroup(f.ctx(), &group.QueryGroupPoliciesByGroupRequest{GroupId: 1})
	require.NoError(t, err)
	require.Len(t, res.GroupPolicies, 1)
	f.groupID = 1
	f.policy, err = sdk.AccAddressFromBech32(res.GroupPolicies[0].Address)
	require.NoError(t, err)
	f.requireTx(f.member, "", banktypes.NewMsgSend(f.member.addr, f.policy, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000))))

	// the genesis authority hands the role to the group policy and names
	// the operator with scope "*"
	f.requireTx(f.member, "", &safetytypes.MsgUpdateParams{
		Authority: f.member.String(),
		Params:    safetytypes.Params{Authority: f.policy.String(), Operator: f.operator.String(), Scope: []string{safetytypes.Wildcard}},
	})
	return f
}

// ctx returns a read-only context on the last committed state.
func (f *groupFixture) ctx() sdk.Context {
	return sdk.NewContext(f.app.CommitMultiStore().CacheMultiStore(),
		cmtproto.Header{Height: f.app.LastBlockHeight(), Time: f.now}, false, log.NewNopLogger())
}

func (f *groupFixture) finalize(txs ...[]byte) []*abci.ExecTxResult {
	f.t.Helper()
	f.now = f.now.Add(5 * time.Second)
	res, err := f.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: f.app.LastBlockHeight() + 1, Time: f.now, Txs: txs})
	require.NoError(f.t, err)
	_, err = f.app.Commit()
	require.NoError(f.t, err)
	require.Len(f.t, res.TxResults, len(txs))
	return res.TxResults
}

// sign encodes msgs as a tx signed by a at its committed sequence.
func (f *groupFixture) sign(a account, msgs ...sdk.Msg) []byte {
	f.t.Helper()
	acc := f.accountKeeper.GetAccount(f.ctx(), a.addr)
	require.NotNil(f.t, acc)
	tx, err := simtestutil.GenSignedMockTx(rand.New(rand.NewSource(1)), f.txConfig, msgs, nil, simtestutil.DefaultGenTxGas, "",
		[]uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, a.priv)
	require.NoError(f.t, err)
	bz, err := f.txConfig.TxEncoder()(tx)
	require.NoError(f.t, err)
	return bz
}

// requireTx requires a tx with msgs, signed by a, to pass CheckTx and
// execution, or with blockedURL set, to fail both on that type URL. A tx
// CheckTx rejects is still delivered, since the decorator stands aside in
// FinalizeBlock and the router must refuse it on its own.
func (f *groupFixture) requireTx(a account, blockedURL string, msgs ...sdk.Msg) *abci.ExecTxResult {
	f.t.Helper()
	bz := f.sign(a, msgs...)
	cres, err := f.app.CheckTx(&abci.RequestCheckTx{Tx: bz, Type: abci.CheckTxType_New})
	require.NoError(f.t, err)
	res := f.finalize(bz)[0]
	if blockedURL == "" {
		require.Zero(f.t, cres.Code, cres.Log)
		require.Zero(f.t, res.Code, res.Log)
		return res
	}
	for _, r := range []struct {
		code      uint32
		codespace string
		log       string
	}{{cres.Code, cres.Codespace, cres.Log}, {res.Code, res.Codespace, res.Log}} {
		require.Equal(f.t, safetytypes.ErrMsgDisabled.ABCICode(), r.code, r.log)
		require.Equal(f.t, safetytypes.ErrMsgDisabled.Codespace(), r.codespace, r.log)
		require.Contains(f.t, r.log, blockedURL)
	}
	return res
}

// requireExecRejected requires a tx with msgs, signed by a, to pass CheckTx
// and fail execution on blockedURL. The decorator checks only the msgs a tx
// lists, so a msg server that another msg server calls directly refuses the
// tx only in FinalizeBlock.
func (f *groupFixture) requireExecRejected(a account, blockedURL string, msgs ...sdk.Msg) {
	f.t.Helper()
	bz := f.sign(a, msgs...)
	cres, err := f.app.CheckTx(&abci.RequestCheckTx{Tx: bz, Type: abci.CheckTxType_New})
	require.NoError(f.t, err)
	require.Zero(f.t, cres.Code, cres.Log)
	res := f.finalize(bz)[0]
	require.Equal(f.t, safetytypes.ErrMsgDisabled.ABCICode(), res.Code, res.Log)
	require.Equal(f.t, safetytypes.ErrMsgDisabled.Codespace(), res.Codespace, res.Log)
	require.Contains(f.t, res.Log, blockedURL)
}

// requireCheckTxRejected requires CheckTx to reject a tx with msgs on
// blockedURL.
func (f *groupFixture) requireCheckTxRejected(a account, blockedURL string, msgs ...sdk.Msg) {
	f.t.Helper()
	cres, err := f.app.CheckTx(&abci.RequestCheckTx{Tx: f.sign(a, msgs...), Type: abci.CheckTxType_New})
	require.NoError(f.t, err)
	require.Equal(f.t, safetytypes.ErrMsgDisabled.ABCICode(), cres.Code, cres.Log)
	require.Contains(f.t, cres.Log, blockedURL)
}

func (f *groupFixture) trip(patterns ...string) {
	f.t.Helper()
	f.requireTx(f.operator, "", safetytypes.NewMsgTrip(f.operator.String(), patterns...))
}

func (f *groupFixture) reset(patterns ...string) {
	f.t.Helper()
	f.requireTx(f.operator, "", safetytypes.NewMsgReset(f.operator.String(), patterns...))
}

func (f *groupFixture) disabled() []string {
	f.t.Helper()
	var out []string
	require.NoError(f.t, f.safetyKeeper.DisabledList.Walk(f.ctx(), nil, func(p string) (bool, error) {
		out = append(out, p)
		return false, nil
	}))
	return out
}

// proposal is a group proposal from the policy, proposed by the member.
func (f *groupFixture) proposal(msgs ...sdk.Msg) *group.MsgSubmitProposal {
	f.t.Helper()
	m, err := group.NewMsgSubmitProposal(f.policy.String(), []string{f.member.String()}, msgs, "", group.Exec_EXEC_UNSPECIFIED, "safety", "safety")
	require.NoError(f.t, err)
	return m
}

// submit submits a group proposal of msgs and returns its id.
func (f *groupFixture) submit(msgs ...sdk.Msg) uint64 {
	f.t.Helper()
	f.requireTx(f.member, "", f.proposal(msgs...))
	f.proposals++
	return f.proposals
}

func (f *groupFixture) vote(id uint64) *group.MsgVote {
	return &group.MsgVote{ProposalId: id, Voter: f.member.String(), Option: group.VOTE_OPTION_YES}
}

func (f *groupFixture) exec(id uint64) *group.MsgExec {
	return &group.MsgExec{ProposalId: id, Executor: f.member.String()}
}

func (f *groupFixture) executorResult(id uint64) group.ProposalExecutorResult {
	f.t.Helper()
	res, err := f.groupKeeper.Proposal(f.ctx(), &group.QueryProposalRequest{ProposalId: id})
	require.NoError(f.t, err)
	return res.Proposal.ExecutorResult
}

func (f *groupFixture) groupMetadata() string {
	f.t.Helper()
	res, err := f.groupKeeper.GroupInfo(f.ctx(), &group.QueryGroupInfoRequest{GroupId: f.groupID})
	require.NoError(f.t, err)
	return res.Info.Metadata
}

func (f *groupFixture) params() safetytypes.Params {
	f.t.Helper()
	p, err := f.safetyKeeper.Params.Get(f.ctx())
	require.NoError(f.t, err)
	return p
}
