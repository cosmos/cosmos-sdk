package simapp

import (
	"bytes"
	"context"
	appmanager "cosmossdk.io/core/app"
	"cosmossdk.io/core/appmodule/v2"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/comet"
	corecontext "cosmossdk.io/core/context"
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
	"github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

type T = transaction.Tx
type HasWeightedOperationsX interface {
	WeightedOperationsX(weight simsx.WeightSource, reg simsx.Registry)
}

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

	appStateFn := simtestutil.AppStateFnY(
		app.AppCodec(),
		app.AuthKeeper.AddressCodec(),
		app.StakingKeeper.ValidatorAddressCodec(),
		toSimsModule(app.ModuleManager().Modules()),
		app.DefaultGenesis(),
	)
	r := rand.New(rand.NewSource(tCfg.Seed))
	//params := simulation.RandomParams(r)
	accounts := slices.DeleteFunc(simtypes.RandomAccounts(r, 5000), // params.NumKeys()),
		func(acc simtypes.Account) bool { // remove blocked accounts
			return app.BankKeeper.GetBlockedAddresses()[acc.AddressBech32]
		})

	appState, accounts, chainID, genesisTimestamp := appStateFn(r, accounts, tCfg)

	appStore := app.GetStore().(cometbfttypes.Store)
	genesisReq := &appmanager.BlockRequest[T]{
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
	activeValidatorSet := NewWeightedValidators().Update(initRsp.ValidatorUpdates)
	valsetHistory := NewValSetHistory(150) // todo: configure
	valsetHistory.Add(genesisReq.Time, activeValidatorSet)
	changeSet, err := genesisState.GetStateChanges()
	require.NoError(t, err)
	stateRoot, err := appStore.Commit(&store.Changeset{Changes: changeSet})
	require.NoError(t, err)

	// next add a block
	emptySimParams := make(map[string]json.RawMessage, 0) // todo read sims params from disk as before
	weights := simsx.ParamWeightSource(emptySimParams)

	factoryRegistry := make(SimsV2Reg, 0)
	// register all msg factories
	for name, v := range app.ModuleManager().Modules() {
		if name == "authz" { //|| // todo: enable when router issue is solved with `/` prefix in MsgTypeURL
			//	name == "slashing" { // todo: enable when tree issue fixed
			continue
		}
		if w, ok := v.(HasWeightedOperationsX); ok {
			w.WeightedOperationsX(weights, &factoryRegistry)
		}
	}
	// todo: register legacy and v1 msg proposals

	const ( // todo: read from CLI instead
		numBlocks     = 500 // 500 default
		maxTXPerBlock = 200 // 200 default
	)

	rootReporter := simsx.NewBasicSimulationReporter()
	blockTime := genesisTimestamp
	var (
		txSkippedCounter int
		txTotalCounter   int
	)
	for i := uint64(0); i < numBlocks; i++ {
		if len(activeValidatorSet) == 0 {
			t.Skipf("run out of validators in block: %d\n", i+1)
			return
		}
		blockTime = blockTime.Add(time.Duration(minTimePerBlock) * time.Second)
		blockTime = blockTime.Add(time.Duration(int64(r.Intn(int(timeRangePerBlock)))) * time.Second)
		valsetHistory.Add(blockTime, activeValidatorSet)
		blockReqN := &appmanager.BlockRequest[T]{
			Height:  genesisReq.Height + 2 + i,
			Time:    blockTime,
			Hash:    stateRoot,
			AppHash: stateRoot,
			ChainId: chainID,
		}
		nextProposer := activeValidatorSet.NextProposer(r)
		fmt.Printf("proposer: %X\n", nextProposer)
		cometInfo := comet.Info{
			//Evidence:        valsetHistory.MissBehaviour(r),
			ProposerAddress: nextProposer,
			LastCommit:      activeValidatorSet.NewCommitInfo(r),
		}

		msgFactoriesFn := factoryRegistry.NextFactoryFn(r)
		ctx = context.WithValue(ctx, corecontext.CometInfoKey, cometInfo) // required for ContextAwareCometInfoService
		resultHandlers := make([]simsx.SimDeliveryResultHandler, 0, maxTXPerBlock)
		var txPerBlockCounter, txPerBlockSkippedCounter int
		blockRsp, updates, err := app.DeliverSims(ctx, blockReqN, func(ctx context.Context) (T, bool) {
			testData := simsx.NewChainDataSource(ctx, r, app.AuthKeeper, app.BankKeeper, app.txConfig.SigningContext().AddressCodec(), accounts...)
			for txPerBlockCounter < maxTXPerBlock {
				txPerBlockCounter++
				msgFactory := msgFactoriesFn()
				reporter := rootReporter.WithScope(msgFactory.MsgType())

				// the stf context is required to access state via keepers
				signers, msg := msgFactory.Create()(ctx, testData, reporter)
				if reporter.IsSkipped() {
					txPerBlockSkippedCounter++
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
		txSkippedCounter += txPerBlockSkippedCounter
		activeValidatorSet = activeValidatorSet.Update(blockRsp.ValidatorUpdates)
		fmt.Printf("active validators: %d at next height %d\n", len(activeValidatorSet), blockReqN.Height)
	}
	fmt.Println("+++ reporter: " + rootReporter.Summary().String())
	fmt.Printf("Tx total: %d skipped: %d\n", txTotalCounter, txSkippedCounter)
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

type weightedFactory struct {
	weight  uint32
	factory simsx.SimMsgFactoryX
}

var _ simsx.Registry = &SimsV2Reg{}

type SimsV2Reg []weightedFactory

func (s *SimsV2Reg) Add(weight uint32, f simsx.SimMsgFactoryX) {
	*s = append(*s, weightedFactory{weight: weight, factory: f})
	slices.SortFunc(*s, func(a, b weightedFactory) int { // sort to make deterministic
		return strings.Compare(sdk.MsgTypeURL(a.factory.MsgType()), sdk.MsgTypeURL(b.factory.MsgType()))
	})
}

func (s SimsV2Reg) NextFactoryFn(r *rand.Rand) func() simsx.SimMsgFactoryX {
	var totalWeight int
	for _, k := range s {
		totalWeight += int(k.weight)
	}
	if totalWeight == 0 {
		panic("registry is empty")
	}
	return func() simsx.SimMsgFactoryX {
		// this is copied from old sims WeightedOperations.getSelectOpFn
		x := r.Intn(totalWeight)
		for i := 0; i < len(s); i++ {
			if x <= int(s[i].weight) {
				return s[i].factory
			}
			x -= int(s[i].weight)
		}
		// shouldn't happen
		return s[0].factory
	}
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

// WeightedValidators is a sorted slice of weighted validators
type WeightedValidators []WeightedValidator

// NewWeightedValidators constructor
func NewWeightedValidators(vals ...WeightedValidator) WeightedValidators {
	r := WeightedValidators(slices.DeleteFunc(vals, func(v WeightedValidator) bool { return v.power == 0 }))
	r.sort()
	return r
}

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
		switch pos, ok := newValset.FindPosition(u.addr); {
		case !ok && u.power == 0: // ignore
			fmt.Printf("ignoring validator: %X -> %d\n", u.addr, u.power)
		case !ok && u.power > 0:
			// add to valset
			newValset = append(newValset, WeightedValidator{addr: u.addr, power: u.power})
			fmt.Printf("adding validator: %X -> %d\n", u.addr, u.power)
		case ok && u.power == 0:
			// remove from valset
			newValset = append(newValset[0:pos], newValset[pos+1:]...)
			fmt.Printf("removing validator: %X\n", u.addr)
		case ok && u.power > 0:
			// update entry
			newValset[pos].power = u.power
			fmt.Printf("update validator: %X -> %d\n", u.addr, u.power)
		default:
			panic(fmt.Sprintf("unsupported validator update: power %d, existing %v", u.power, ok))
		}
	}
	// sort vals by power, addr desc
	newValset.sort()
	return newValset
}

func (v WeightedValidators) FindPosition(addr []byte) (pos int, exists bool) {
	return slices.BinarySearchFunc(v, WeightedValidator{addr: addr}, func(s, o WeightedValidator) int { return bytes.Compare(s.addr, o.addr) })
}

// sort by address asc
func (v *WeightedValidators) sort() {
	slices.SortFunc(*v, func(a, b WeightedValidator) int { return bytes.Compare(a.addr, b.addr) })
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

// NextProposer return a random validator address
func (v WeightedValidators) NextProposer(r *rand.Rand) []byte {
	for i := 0; i < 10; i++ {
		if v := simsx.OneOf(r, v); !v.Offline {
			return v.addr
		}
	}
	panic("no proposer found in 10 tries")
}

type WeightedValidator struct {
	power   int64
	addr    []byte
	Offline bool
}

type historicValSet struct {
	blockTime time.Time
	vals      WeightedValidators
}
type ValSetHistory struct {
	maxElements int
	blockOffset int
	historic    []historicValSet
}

func NewValSetHistory(maxElements int) *ValSetHistory {
	return &ValSetHistory{
		maxElements: maxElements,
		blockOffset: 1, // start at height 1
		historic:    make([]historicValSet, 0, maxElements),
	}
}

func (h *ValSetHistory) Add(blockTime time.Time, vals WeightedValidators) {
	newEntry := historicValSet{blockTime: blockTime, vals: vals}
	if len(h.historic) >= h.maxElements {
		h.historic = append(h.historic[1:], newEntry)
		h.blockOffset++
		return
	}
	h.historic = append(h.historic, newEntry)
}

func (h *ValSetHistory) MissBehaviour(r *rand.Rand) []comet.Evidence {
	if r.Intn(1000) != 0 { // 0.1% chance
		return nil
	}
	n := r.Intn(len(h.historic))
	badVal := simsx.OneOf(r, h.historic[n].vals)
	evidence := comet.Evidence{
		Type:             comet.DuplicateVote,
		Validator:        comet.Validator{Address: badVal.addr, Power: badVal.power},
		Height:           int64(h.blockOffset + n),
		Time:             h.historic[n].blockTime,
		TotalVotingPower: h.historic[n].vals.TotalPower(), // historic
		//TotalVotingPower: h.vals[len(h.vals)-1].vals.TotalPower(), // current
	}
	if otherEvidence := h.MissBehaviour(r); otherEvidence != nil {
		return append([]comet.Evidence{evidence}, otherEvidence...)
	}
	return []comet.Evidence{evidence}
}
