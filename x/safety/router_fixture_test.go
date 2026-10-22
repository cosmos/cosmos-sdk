package safety_test

import (
	"fmt"
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
	upgrademodulev1 "cosmossdk.io/api/cosmos/upgrade/module/v1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appconfig"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log/v2"
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/x/auth" // import as blank for app wiring
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import as blank for app wiring
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	_ "github.com/cosmos/cosmos-sdk/x/authz/module" // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/bank"         // import as blank for app wiring
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	_ "github.com/cosmos/cosmos-sdk/x/consensus" // import as blank for app wiring
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/distribution" // import as blank for app wiring
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/gov" // import as blank for app wiring
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/safety"
	safetyante "github.com/cosmos/cosmos-sdk/x/safety/ante"
	"github.com/cosmos/cosmos-sdk/x/safety/keeper"
	"github.com/cosmos/cosmos-sdk/x/safety/types"
	_ "github.com/cosmos/cosmos-sdk/x/staking" // import as blank for app wiring
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/upgrade" // import as blank for app wiring
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
)

// testAccount is a funded genesis account that signs txs.
type testAccount struct {
	priv cryptotypes.PrivKey
	addr sdk.AccAddress
}

func newTestAccount() testAccount {
	priv := secp256k1.GenPrivKey()
	return testAccount{priv: priv, addr: sdk.AccAddress(priv.PubKey().Address())}
}

func (a testAccount) String() string { return a.addr.String() }

// routerFixture is a full app built through depinject with real auth, bank,
// staking, distribution, gov, authz, consensus and upgrade modules, with
// x/safety wired by hand on top, as a chain wires it. x/safety installs its
// keeper as the BaseApp circuit breaker, so every msg
// that goes through the MsgServiceRouter, whether top level, nested in authz
// MsgExec, or executed by a gov proposal, hits the x/safety check.
//
// With withAnte false, the app runs the default x/auth ante chain and the
// router is the only x/safety gate. With withAnte true, the ante chain is the
// SimApp chain with SafetyDecorator placed before fee deduction.
type routerFixture struct {
	t *testing.T

	app      *runtime.App
	txConfig client.TxConfig
	registry codectypes.InterfaceRegistry

	accountKeeper   authkeeper.AccountKeeper
	bankKeeper      bankkeeper.Keeper
	stakingKeeper   *stakingkeeper.Keeper
	distrKeeper     distrkeeper.Keeper
	govKeeper       *govkeeper.Keeper
	upgradeKeeper   *upgradekeeper.Keeper
	consensusKeeper consensuskeeper.Keeper
	safetyKeeper    keeper.Keeper

	// user is the first genesis account, so it owns the genesis delegation and
	// decides every gov vote. The fixture makes it the x/safety operator with
	// scope "*" through a gov proposal.
	user testAccount
	// grantee is a second funded account, used as the authz grantee.
	grantee testAccount

	gov       string
	blockTime time.Time
	rng       *rand.Rand
}

// txModuleSkipAnte wires x/auth/tx without its default ante handler, so the
// test can install its own chain.
func txModuleSkipAnte() configurator.ModuleOption {
	return func(config *configurator.Config) {
		config.ModuleConfigs["tx"] = &appv1alpha1.ModuleConfig{
			Name:   "tx",
			Config: appconfig.WrapAny(&txconfigv1.Config{SkipAnteHandler: true}),
		}
	}
}

func upgradeModule() configurator.ModuleOption {
	return func(config *configurator.Config) {
		config.ModuleConfigs["upgrade"] = &appv1alpha1.ModuleConfig{
			Name:   "upgrade",
			Config: appconfig.WrapAny(&upgrademodulev1.Module{}),
		}
	}
}

// wireSafety adds x/safety to an app built through depinject, before it
// loads: it mounts the store, builds the keeper with the given authority,
// installs the keeper as the circuit breaker, registers the module and its
// services, and sets the init and export genesis order to initGenesis followed
// by x/safety. The app config must omit the init genesis order.
func wireSafety(t *testing.T, app *runtime.App, cdc codec.Codec, ac address.Codec, authority string, initGenesis ...string) keeper.Keeper {
	t.Helper()
	key := storetypes.NewKVStoreKey(types.StoreKey)
	require.NoError(t, app.RegisterStores(key))
	k := keeper.NewKeeper(cdc, runtime.NewKVStoreService(key), ac, cdc.InterfaceRegistry(), authority)
	app.SetCircuitBreaker(k)
	require.NoError(t, app.RegisterModules(safety.NewAppModule(k)))

	order := append(initGenesis, types.ModuleName)
	app.ModuleManager.SetOrderInitGenesis(order...)
	app.ModuleManager.SetOrderExportGenesis(order...)
	app.SetInitChainer(app.InitChainer)
	return k
}

// safetyAnteHandler mirrors the SimApp ante chain: SafetyDecorator right
// after context setup, ahead of fee deduction and sequence increment.
func safetyAnteHandler(ak authkeeper.AccountKeeper, bk bankkeeper.Keeper, sk keeper.Keeper, txConfig client.TxConfig) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		authante.NewSetUpContextDecorator(),
		safetyante.NewSafetyDecorator(sk),
		authante.NewExtensionOptionsDecorator(nil),
		authante.NewValidateBasicDecorator(),
		authante.NewTxTimeoutHeightDecorator(),
		authante.NewValidateMemoDecorator(ak),
		authante.NewConsumeGasForTxSizeDecorator(ak),
		authante.NewDeductFeeDecorator(ak, bk, nil, nil),
		authante.NewSetPubKeyDecorator(ak),
		authante.NewValidateSigCountDecorator(ak),
		authante.NewSigGasConsumeDecorator(ak, authante.DefaultSigVerificationGasConsumer),
		authante.NewSigVerificationDecorator(ak, txConfig.SignModeHandler()),
		authante.NewIncrementSequenceDecorator(ak),
	)
}

func newRouterFixture(t *testing.T, withAnte bool) *routerFixture {
	t.Helper()

	f := &routerFixture{
		t:         t,
		user:      newTestAccount(),
		grantee:   newTestAccount(),
		gov:       authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		blockTime: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		rng:       rand.New(rand.NewSource(1)),
	}

	funds := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(1_000_000_000_000)))
	genesisAccounts := []sims.GenesisAccount{
		{GenesisAccount: authtypes.NewBaseAccount(f.user.addr, nil, 0, 0), Coins: funds},
		{GenesisAccount: authtypes.NewBaseAccount(f.grantee.addr, nil, 1, 0), Coins: funds},
	}

	txModule := configurator.TxModule()
	if withAnte {
		txModule = txModuleSkipAnte()
	}

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
				configurator.DistributionModule(),
				configurator.GovModule(),
				configurator.AuthzModule(),
				configurator.ConsensusModule(),
				upgradeModule(),
				txModule,
				// x/safety is wired by hand below and has no app config
				configurator.OmitInitGenesis(),
			),
			depinject.Supply(log.NewNopLogger()),
		),
		&appBuilder, &cdc,
		&f.txConfig, &f.registry, &f.accountKeeper, &f.bankKeeper, &f.stakingKeeper, &f.distrKeeper,
		&f.govKeeper, &f.upgradeKeeper, &f.consensusKeeper,
	))
	app := appBuilder.Build(dbm.NewMemDB())
	f.safetyKeeper = wireSafety(t, app, cdc, f.accountKeeper.AddressCodec(), f.gov,
		"auth", "bank", "distribution", "staking", "gov", "authz", "consensus", "upgrade")
	if withAnte {
		app.SetAnteHandler(safetyAnteHandler(f.accountKeeper, f.bankKeeper, f.safetyKeeper, f.txConfig))
	}
	require.NoError(t, app.Load(true))
	f.app = app

	valSet, err := sims.CreateRandomValidatorSet()
	require.NoError(t, err)
	var (
		genAccs  []authtypes.GenesisAccount
		balances []banktypes.Balance
	)
	for _, ga := range genesisAccounts {
		genAccs = append(genAccs, ga.GenesisAccount)
		balances = append(balances, banktypes.Balance{Address: ga.GenesisAccount.GetAddress().String(), Coins: ga.Coins})
	}
	genesis, err := sims.GenesisStateWithValSet(cdc, app.DefaultGenesis(), valSet, genAccs, balances...)
	require.NoError(t, err)
	stateBytes, err := cmtjson.MarshalIndent(genesis, "", " ")
	require.NoError(t, err)
	_, err = app.InitChain(&abci.RequestInitChain{
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: sims.DefaultConsensusParams,
		AppStateBytes:   stateBytes,
	})
	require.NoError(t, err)
	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             app.LastBlockHeight() + 1,
		NextValidatorsHash: valSet.Hash(),
	})
	require.NoError(t, err)

	// commit block 1
	_, err = app.Commit()
	require.NoError(t, err)

	params, err := f.safetyKeeper.Params.Get(f.ctx())
	require.NoError(t, err)
	require.Equal(t, types.DefaultParams(f.gov), params)

	// make user the operator with scope "*" the way a chain would: through a
	// gov proposal that executes x/safety MsgUpdateParams
	id := f.submitProposal(&types.MsgUpdateParams{
		Authority: f.gov,
		Params:    types.Params{Authority: f.gov, Operator: f.user.String(), Scope: []string{types.Wildcard}},
	})
	f.endVoting()
	f.requireProposalStatus(id, govv1.StatusPassed)

	params, err = f.safetyKeeper.Params.Get(f.ctx())
	require.NoError(t, err)
	require.Equal(t, f.user.String(), params.Operator)
	return f
}

// ctx returns a read-only context on the last committed state.
func (f *routerFixture) ctx() sdk.Context {
	return sdk.NewContext(
		f.app.CommitMultiStore().CacheMultiStore(),
		cmtproto.Header{Height: f.app.LastBlockHeight(), Time: f.blockTime},
		false, log.NewNopLogger(),
	)
}

// checkCtx returns the CheckTx state, which reflects every tx CheckTx admitted
// since the last commit, including the fees their ante handlers deducted.
func (f *routerFixture) checkCtx() sdk.Context { return f.app.NewContext(true) }

// seq returns the account's committed sequence.
func (f *routerFixture) seq(a testAccount) uint64 {
	acc := f.accountKeeper.GetAccount(f.ctx(), a.addr)
	require.NotNil(f.t, acc)
	return acc.GetSequence()
}

func (f *routerFixture) accNum(a testAccount) uint64 {
	acc := f.accountKeeper.GetAccount(f.ctx(), a.addr)
	require.NotNil(f.t, acc)
	return acc.GetAccountNumber()
}

// sign returns an encoded tx of msgs, signed by a at sequence seq.
func (f *routerFixture) sign(a testAccount, seq uint64, fee sdk.Coins, msgs ...sdk.Msg) []byte {
	f.t.Helper()
	tx, err := sims.GenSignedMockTx(f.rng, f.txConfig, msgs, fee, sims.DefaultGenTxGas, "",
		[]uint64{f.accNum(a)}, []uint64{seq}, a.priv)
	require.NoError(f.t, err)
	bz, err := f.txConfig.TxEncoder()(tx)
	require.NoError(f.t, err)
	return bz
}

// signNext signs msgs at a's committed sequence.
func (f *routerFixture) signNext(a testAccount, fee sdk.Coins, msgs ...sdk.Msg) []byte {
	return f.sign(a, f.seq(a), fee, msgs...)
}

// deliverAt runs FinalizeBlock with txs at the given block time and commits.
func (f *routerFixture) deliverAt(blockTime time.Time, txs ...[]byte) []*abci.ExecTxResult {
	f.t.Helper()
	f.blockTime = blockTime
	res, err := f.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: f.app.LastBlockHeight() + 1,
		Time:   f.blockTime,
		Txs:    txs,
	})
	require.NoError(f.t, err)
	_, err = f.app.Commit()
	require.NoError(f.t, err)
	require.Len(f.t, res.TxResults, len(txs))
	return res.TxResults
}

// deliver runs one block one second after the last one.
func (f *routerFixture) deliver(txs ...[]byte) []*abci.ExecTxResult {
	f.t.Helper()
	return f.deliverAt(f.blockTime.Add(time.Second), txs...)
}

// deliverOne runs a block with a single tx signed by a and returns its result.
func (f *routerFixture) deliverOne(a testAccount, msgs ...sdk.Msg) *abci.ExecTxResult {
	f.t.Helper()
	return f.deliver(f.signNext(a, nil, msgs...))[0]
}

func (f *routerFixture) checkTx(bz []byte) *abci.ResponseCheckTx {
	f.t.Helper()
	res, err := f.app.CheckTx(&abci.RequestCheckTx{Tx: bz, Type: abci.CheckTxType_New})
	require.NoError(f.t, err)
	return res
}

// trip trips patterns as the operator and requires success.
func (f *routerFixture) trip(patterns ...string) {
	f.t.Helper()
	res := f.deliverOne(f.user, types.NewMsgTrip(f.user.String(), patterns...))
	require.Zero(f.t, res.Code, res.Log)
}

// reset resets patterns as the operator and requires success.
func (f *routerFixture) reset(patterns ...string) {
	f.t.Helper()
	res := f.deliverOne(f.user, types.NewMsgReset(f.user.String(), patterns...))
	require.Zero(f.t, res.Code, res.Log)
}

// submitProposal submits a gov proposal of msgs with the minimum deposit, so it
// enters voting at once, and votes yes with the only delegator in the same
// block. It returns the proposal id.
func (f *routerFixture) submitProposal(msgs ...sdk.Msg) uint64 {
	f.t.Helper()
	ctx := f.ctx()
	id, err := f.govKeeper.ProposalID.Peek(ctx)
	require.NoError(f.t, err)
	params, err := f.govKeeper.Params.Get(ctx)
	require.NoError(f.t, err)

	submit, err := govv1.NewMsgSubmitProposal(msgs, params.MinDeposit, f.user.String(), "", "x/safety test", "x/safety test", false)
	require.NoError(f.t, err)
	vote := govv1.NewMsgVote(f.user.addr, id, govv1.OptionYes, "")

	seq := f.seq(f.user)
	res := f.deliver(f.sign(f.user, seq, nil, submit), f.sign(f.user, seq+1, nil, vote))
	require.Zero(f.t, res[0].Code, "submit: %s", res[0].Log)
	require.Zero(f.t, res[1].Code, "vote: %s", res[1].Log)
	f.requireProposalStatus(id, govv1.StatusVotingPeriod)
	return id
}

// endVoting runs an empty block past the voting period, so gov tallies and
// executes every proposal in voting.
func (f *routerFixture) endVoting() {
	f.t.Helper()
	params, err := f.govKeeper.Params.Get(f.ctx())
	require.NoError(f.t, err)
	f.deliverAt(f.blockTime.Add(*params.VotingPeriod + time.Hour))
}

func (f *routerFixture) proposal(id uint64) govv1.Proposal {
	f.t.Helper()
	p, err := f.govKeeper.Proposals.Get(f.ctx(), id)
	require.NoError(f.t, err)
	return p
}

func (f *routerFixture) requireProposalStatus(id uint64, status govv1.ProposalStatus) govv1.Proposal {
	f.t.Helper()
	p := f.proposal(id)
	require.Equal(f.t, status, p.Status, "proposal %d failed reason: %q", id, p.FailedReason)
	return p
}

func (f *routerFixture) balance(ctx sdk.Context, a testAccount) sdkmath.Int {
	return f.bankKeeper.GetBalance(ctx, a.addr, sdk.DefaultBondDenom).Amount
}

func (f *routerFixture) validator() sdk.ValAddress {
	f.t.Helper()
	vals, err := f.stakingKeeper.GetAllValidators(f.ctx())
	require.NoError(f.t, err)
	require.NotEmpty(f.t, vals)
	bz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(vals[0].GetOperator())
	require.NoError(f.t, err)
	return bz
}

// callHandler runs a router handler on a branch of ctx. A panic from the
// module handler counts as passing the breaker, since the breaker check runs
// before the module handler.
func callHandler(ctx sdk.Context, handler baseapp.MsgServiceHandler, msg sdk.Msg) (err error) {
	cacheCtx, _ := ctx.CacheContext()
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("handler panicked: %v", r)
		}
	}()
	_, err = handler(cacheCtx, msg)
	return err
}
