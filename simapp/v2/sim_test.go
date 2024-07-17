package simapp

import (
	"bytes"
	"context"
	"cosmossdk.io/core/appmodule/v2"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/comet"
	corecontext "cosmossdk.io/core/context"
	"cosmossdk.io/core/server"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	cometbfttypes "cosmossdk.io/server/v2/cometbft/types"
	consensustypes "cosmossdk.io/x/consensus/types"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simsx"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	"github.com/huandu/skiplist"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
	"iter"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"testing"
	"time"
)

type T = transaction.Tx
type (
	HasWeightedOperationsX interface {
		WeightedOperationsX(weight simsx.WeightSource, reg simsx.Registry)
	}
	HasWeightedOperationsXWithProposals interface {
		WeightedOperationsX(weights simsx.WeightSource, reg simsx.Registry, proposals iter.Seq2[uint32, simsx.SimMsgFactoryX],
			legacyProposals []simtypes.WeightedProposalContent) //nolint: staticcheck // used for legacy proposal types
	}
	HasProposalMsgsX interface {
		ProposalMsgsX(weights simsx.WeightSource, reg simsx.Registry)
	}
)

const (
	minTimePerBlock int64 = 10000 / 2

	maxTimePerBlock int64 = 10000

	timeRangePerBlock = maxTimePerBlock - minTimePerBlock
)

func TestSimsAppV2(t *testing.T) {
	DefaultNodeHome = t.TempDir()
	currentDir, err := os.Getwd()
	require.NoError(t, err)
	configPath := filepath.Join(currentDir, "testdata")
	v, err := serverv2.ReadConfig(configPath)
	require.NoError(t, err)
	v.Set("home", DefaultNodeHome)
	//v.Set("store.app-db-backend", "memdb") // todo: I had added this new type to speed up testing. Does it make sense this way?
	logger := log.NewTestLoggerInfo(t)
	app := NewSimApp[T](logger, v)

	tCfg := cli.NewConfigFromFlags().With(t, 1, nil)

	appStateFn := simtestutil.AppStateFn(
		app.AppCodec(),
		app.AuthKeeper.AddressCodec(),
		app.StakingKeeper.ValidatorAddressCodec(),
		toSimsModule(app.ModuleManager().Modules()),
		app.DefaultGenesis(),
	)
	r := rand.New(rand.NewSource(tCfg.Seed))
	params := simulation.RandomParams(r)
	accounts := slices.DeleteFunc(simtypes.RandomAccounts(r, params.NumKeys()),
		func(acc simtypes.Account) bool { // remove blocked accounts
			return app.BankKeeper.GetBlockedAddresses()[acc.AddressBech32]
		})

	appState, accounts, chainID, genesisTimestamp := appStateFn(r, accounts, tCfg)

	appStore := app.GetStore().(cometbfttypes.Store)
	genesisReq := &server.BlockRequest[T]{
		Height:    0, // todo: or 1?
		Time:      genesisTimestamp,
		Hash:      make([]byte, 32),
		ChainId:   chainID,
		AppHash:   make([]byte, 32),
		IsGenesis: true,
	}
	ctx, done := context.WithCancel(context.Background())
	defer done()
	genesisCtx := context.WithValue(ctx, corecontext.InitInfoKey, &consensustypes.MsgUpdateParams{
		Authority: app.GetConsensusAuthority(), // todo: what else is needed in setup ?
		Block: &cmtproto.BlockParams{
			MaxBytes: 200000,
			MaxGas:   100_000_000,
		},
		Evidence: &cmtproto.EvidenceParams{
			MaxAgeNumBlocks: 302400,
			MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
			MaxBytes:        10000,
		},
		Validator: &cmtproto.ValidatorParams{PubKeyTypes: []string{cmttypes.ABCIPubKeyTypeEd25519, cmttypes.ABCIPubKeyTypeSecp256k1}},
	})

	initRsp, genesisState, err := app.InitGenesis(genesisCtx, genesisReq, appState, &genericTxDecoder[T]{txConfig: app.TxConfig()})
	require.NoError(t, err)
	activeValidatorSet := NewValSet().Update(initRsp.ValidatorUpdates)
	valsetHistory := NewValSetHistory(150) // todo: configure
	valsetHistory.Add(genesisReq.Time, activeValidatorSet)
	require.NoError(t, appStore.SetInitialVersion(genesisReq.Height))
	changeSet, err := genesisState.GetStateChanges()
	require.NoError(t, err)
	stateRoot, err := appStore.Commit(&store.Changeset{Changes: changeSet})
	require.NoError(t, err)

	emptySimParams := make(map[string]json.RawMessage) // todo read sims params from disk as before
	weights := simsx.ParamWeightSource(emptySimParams)

	// get all proposal types
	proposalRegistry := make(simsx.SimsV2Reg)
	for _, m := range app.ModuleManager().Modules() {
		switch xm := m.(type) {
		case HasProposalMsgsX:
			xm.ProposalMsgsX(weights, proposalRegistry)
			// todo: register legacy and v1 msg proposals
		}
	}
	// register all msg factories
	factoryRegistry := make(simsx.SimsV2Reg)
	for _, m := range app.ModuleManager().Modules() {
		switch xm := m.(type) {
		case HasWeightedOperationsX:
			xm.WeightedOperationsX(weights, factoryRegistry)
		case HasWeightedOperationsXWithProposals:
			xm.WeightedOperationsX(weights, factoryRegistry, proposalRegistry.Iterator(), nil)
		}
	}
	msgTypes := maps.Keys(factoryRegistry)
	sort.Strings(msgTypes)
	for _, k := range msgTypes {
		x := factoryRegistry[k]
		fmt.Printf("factory: %d -> %s\n", x.Weight, k)
	}
	const ( // todo: read from CLI instead
		numBlocks     = 1200 // 500 default
		maxTXPerBlock = 650  // 200 default
	)

	rootReporter := simsx.NewBasicSimulationReporter()
	blockTime := genesisTimestamp
	var (
		txSkippedCounter int
		txTotalCounter   int
	)
	futureOpsReg := newFutureOpsRegistry()
	msgFactoriesFn := factoryRegistry.NextFactoryFn(r)

	for i := 0; i < numBlocks; i++ {
		if len(activeValidatorSet) == 0 {
			t.Skipf("run out of validators in block: %d\n", i+1)
			return
		}
		blockTime = blockTime.Add(time.Duration(minTimePerBlock) * time.Second)
		blockTime = blockTime.Add(time.Duration(int64(r.Intn(int(timeRangePerBlock)))) * time.Second)
		valsetHistory.Add(blockTime, activeValidatorSet)
		blockReqN := &server.BlockRequest[T]{
			Height:  uint64(2 + i),
			Time:    blockTime,
			Hash:    stateRoot,
			AppHash: stateRoot,
			ChainId: chainID,
		}
		cometInfo := comet.Info{
			ValidatorsHash:  nil,
			Evidence:        valsetHistory.MissBehaviour(r),
			ProposerAddress: activeValidatorSet[0].addr,
			LastCommit:      activeValidatorSet.NewCommitInfo(r),
		}
		fOps, pos := futureOpsReg.findScheduled(blockTime), 0
		nextFactoryFn := func() simsx.SimMsgFactoryX {
			if pos < len(fOps) {
				pos++
				return fOps[pos-1]
			}
			return msgFactoriesFn()
		}
		ctx = context.WithValue(ctx, corecontext.CometInfoKey, cometInfo) // required for ContextAwareCometInfoService
		resultHandlers := make([]simsx.SimDeliveryResultHandler, 0, maxTXPerBlock)
		var txPerBlockCounter int
		blockRsp, updates, err := app.DeliverSims(ctx, blockReqN, func(ctx context.Context) (T, bool) {
			testData := simsx.NewChainDataSource(ctx, r, app.AuthKeeper, app.BankKeeper, app.txConfig.SigningContext().AddressCodec(), accounts...)
			for txPerBlockCounter < maxTXPerBlock {
				txPerBlockCounter++
				msgFactory := nextFactoryFn()
				reporter := rootReporter.WithScope(msgFactory.MsgType())
				if fx, ok := msgFactory.(simsx.HasFutureOpsRegistry); ok {
					fx.SetFutureOpsRegistry(futureOpsReg)
				}

				// the stf context is required to access state via keepers
				signers, msg := msgFactory.Create()(ctx, testData, reporter)
				if reporter.IsSkipped() {
					txSkippedCounter++
					require.NoError(t, reporter.Close())
					continue
				}
				resultHandlers = append(resultHandlers, msgFactory.DeliveryResultHandler())
				reporter.Success(msg)
				require.NoError(t, reporter.Close())

				tx, err := buildTestTX(ctx, app.AuthKeeper, signers, msg, r, app.txConfig, chainID)
				require.NoError(t, err)
				return tx, false
			}
			return nil, true
		})
		require.NoError(t, err)
		changeSet, err = updates.GetStateChanges()
		require.NoError(t, err)
		stateRoot, err = appStore.Commit(&store.Changeset{Changes: changeSet})
		require.NoError(t, err)
		require.Equal(t, len(resultHandlers), len(blockRsp.TxResults), "txPerBlockCounter: %d, totalSkipped: %d", txPerBlockCounter, txSkippedCounter)
		for i, v := range blockRsp.TxResults {
			require.NoError(t, resultHandlers[i](v.Error))
		}
		txTotalCounter += txPerBlockCounter
		activeValidatorSet = activeValidatorSet.Update(blockRsp.ValidatorUpdates)
		fmt.Printf("active validator set: %d\n", len(activeValidatorSet))
	}
	fmt.Println("+++ reporter:\n" + rootReporter.Summary().String())
	fmt.Printf("Tx total: %d skipped: %d\n", txTotalCounter, txSkippedCounter)
}

type futureOpsRegistry struct {
	list *skiplist.SkipList
}

var _ skiplist.Comparable = timeComparator{}

// used for skiplist
type timeComparator struct {
}

func (t timeComparator) Compare(lhs, rhs interface{}) int {
	return lhs.(time.Time).Compare(rhs.(time.Time))
}

func (t timeComparator) CalcScore(key interface{}) float64 {
	return float64(key.(time.Time).UnixNano())
}

func newFutureOpsRegistry() *futureOpsRegistry {
	return &futureOpsRegistry{list: skiplist.New(skiplist.Int64)}
}

func (l *futureOpsRegistry) Add(blockTime time.Time, fx simsx.SimMsgFactoryX) {
	if fx == nil {
		panic("message factory must not be nil")
	}
	if blockTime.IsZero() {
		return
	}
	var scheduledOps []simsx.SimMsgFactoryX
	if e := l.list.Get(blockTime); e != nil {
		scheduledOps = e.Value.([]simsx.SimMsgFactoryX)
	}
	scheduledOps = append(scheduledOps, fx)
	l.list.Set(blockTime, scheduledOps)
}

func (l *futureOpsRegistry) findScheduled(blockTime time.Time) []simsx.SimMsgFactoryX {
	var r []simsx.SimMsgFactoryX
	for {
		e := l.list.Front()
		if e == nil || e.Key().(time.Time).After(blockTime) {
			break
		}
		r = append(r, e.Value.([]simsx.SimMsgFactoryX)...)
		l.list.RemoveFront()
	}
	return r
}

func buildTestTX(
	ctx context.Context,
	ak simsx.AccountSource,
	senders []simsx.SimAccount,
	msg sdk.Msg,
	r *rand.Rand,
	txGen client.TxConfig,
	chainID string,
) (sdk.Tx, error) {
	accountNumbers := make([]uint64, len(senders))
	sequenceNumbers := make([]uint64, len(senders))
	for i := 0; i < len(senders); i++ {
		acc := ak.GetAccount(ctx, senders[i].Address)
		accountNumbers[i] = acc.GetAccountNumber()
		sequenceNumbers[i] = acc.GetSequence()
	}
	fees := senders[0].LiquidBalance().RandFees()
	return sims.GenSignedMockTx(
		r,
		txGen,
		[]sdk.Msg{msg},
		fees,
		sims.DefaultGenTxGas,
		chainID,
		accountNumbers,
		sequenceNumbers,
		Collect(senders, func(a simsx.SimAccount) cryptotypes.PrivKey { return a.PrivKey })...,
	)
}

func toSimsModule(modules map[string]appmodule.AppModule) []module.AppModuleSimulation {
	r := make([]module.AppModuleSimulation, 0, len(modules))
	names := maps.Keys(modules)
	slices.Sort(names) // make deterministic
	for _, v := range names {
		if m, ok := modules[v].(module.AppModuleSimulation); ok {
			r = append(r, m)
		}
	}
	return r
}

var _ transaction.Codec[transaction.Tx] = &genericTxDecoder[transaction.Tx]{}

// todo: this is the same as in commands
type genericTxDecoder[T transaction.Tx] struct {
	txConfig client.TxConfig
}

// Decode implements transaction.Codec.
func (t *genericTxDecoder[T]) Decode(bz []byte) (T, error) {
	var out T
	tx, err := t.txConfig.TxDecoder()(bz)
	if err != nil {
		return out, err
	}

	var ok bool
	out, ok = tx.(T)
	if !ok {
		return out, errors.New("unexpected Tx type")
	}

	return out, nil
}

// DecodeJSON implements transaction.Codec.
func (t *genericTxDecoder[T]) DecodeJSON(bz []byte) (T, error) {
	var out T
	tx, err := t.txConfig.TxJSONDecoder()(bz)
	if err != nil {
		return out, err
	}

	var ok bool
	out, ok = tx.(T)
	if !ok {
		return out, errors.New("unexpected Tx type")
	}

	return out, nil
}

func Collect[T, E any](source []T, f func(a T) E) []E {
	r := make([]E, len(source))
	for i, v := range source {
		r[i] = f(v)
	}
	return r
}

// NewValSet constructor
func NewValSet() WeightedValidators {
	return make(WeightedValidators, 0)
}

type WeightedValidators []WeightedValidator

func (v WeightedValidators) Update(updates []appmodulev2.ValidatorUpdate) WeightedValidators {
	if len(updates) == 0 {
		return v
	}
	const truncatedSize = 20
	valUpdates := simsx.Collect(updates, func(u appmodulev2.ValidatorUpdate) WeightedValidator {
		hash := sha256.Sum256(u.PubKey)
		return WeightedValidator{power: u.Power, addr: hash[:truncatedSize]}
	})
	newValset := slices.Clone(v)
	for _, u := range valUpdates {
		pos := slices.IndexFunc(newValset, func(val WeightedValidator) bool {
			return bytes.Equal(u.addr, val.addr)
		})
		if pos == -1 {
			if u.power > 0 {
				newValset = append(newValset, u)
			}
			continue
		}
		if u.power == 0 {
			newValset = append(newValset[0:pos], newValset[pos+1:]...)
			continue
		}
		newValset[pos].power = u.power
	}

	// sort vals by power
	slices.SortFunc(newValset, func(a, b WeightedValidator) int {
		switch {
		case a.power < b.power:
			return 1
		case a.power > a.power:
			return -1
		default:
			return bytes.Compare(a.addr, b.addr)
		}
	})
	return newValset
}

// NewCommitInfo build Comet commit info for the validator set
func (v WeightedValidators) NewCommitInfo(r *rand.Rand) comet.CommitInfo {
	// todo: refactor to transition matrix?
	if r.Intn(10) == 0 {
		v[rand.Intn(len(v))].Offline = r.Intn(2) == 0
	}
	votes := make([]comet.VoteInfo, 0, len(v))
	for i := range v {
		if v[i].Offline {
			continue
		}
		votes = append(votes, comet.VoteInfo{
			Validator:   comet.Validator{Address: v[i].addr, Power: v[i].power},
			BlockIDFlag: comet.BlockIDFlagCommit,
		})
	}
	return comet.CommitInfo{Round: int32(r.Uint32()), Votes: votes}
}

func (v WeightedValidators) TotalPower() int64 {
	var r int64
	for _, val := range v {
		r += val.power
	}
	return r
}

type WeightedValidator struct {
	power   int64
	addr    []byte
	Offline bool
}

func must[T any](r T, err error) T {
	if err != nil {
		panic(err)
	}
	return r
}

type historicValSet struct {
	blockTime time.Time
	vals      WeightedValidators
}
type ValSetHistory struct {
	maxElements int
	blockOffset int
	vals        []historicValSet
}

func NewValSetHistory(maxElements int) *ValSetHistory {
	return &ValSetHistory{
		maxElements: maxElements,
		blockOffset: 1, // start at height 1
		vals:        make([]historicValSet, 0, maxElements),
	}
}

func (h *ValSetHistory) Add(blockTime time.Time, vals WeightedValidators) {
	newEntry := historicValSet{blockTime: blockTime, vals: vals}
	if len(h.vals) >= h.maxElements {
		h.vals = append(h.vals[1:], newEntry)
		h.blockOffset++
		return
	}
	h.vals = append(h.vals, newEntry)
}

func (h *ValSetHistory) MissBehaviour(r *rand.Rand) []comet.Evidence {
	if r.Intn(100) != 0 { // 1% chance
		return nil
	}
	n := r.Intn(len(h.vals))
	badVal := simsx.OneOf(r, h.vals[n].vals)
	evidence := comet.Evidence{
		Type:             comet.DuplicateVote,
		Validator:        comet.Validator{Address: badVal.addr, Power: badVal.power},
		Height:           int64(h.blockOffset + n),
		Time:             h.vals[n].blockTime,
		TotalVotingPower: h.vals[n].vals.TotalPower(),
	}
	if otherEvidence := h.MissBehaviour(r); otherEvidence != nil {
		return append([]comet.Evidence{evidence}, otherEvidence...)
	}
	return []comet.Evidence{evidence}
}
