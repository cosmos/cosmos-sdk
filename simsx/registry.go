package simsx

import (
	"context"
	"math/rand"
	"time"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/log"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Registry is an abstract entry point to register message factories with weights
type Registry interface {
	Add(weight uint32, f SimMsgFactoryX)
}
type FutureOpsRegistry interface {
	Add(blockTime time.Time, f SimMsgFactoryX)
}
type AccountSourceX interface {
	AccountSource
	ModuleAccountSource
}

var (
	_ Registry = &WeightedOperationRegistryAdapter{}
	_ Registry = &SimsProposalRegistryAdapter{}
)

// common types for abstract registry without generics
type regCommon struct {
	reporter     SimulationReporter
	ak           AccountSourceX
	bk           BalanceSource
	addressCodec address.Codec
	txConfig     client.TxConfig
	logger       log.Logger
}

func (c regCommon) newChainDataSource(ctx context.Context, r *rand.Rand, accs ...simtypes.Account) *ChainDataSource {
	return NewChainDataSource(ctx, r, c.ak, c.bk, c.addressCodec, accs...)
}

type AbstractRegistry[T any] struct {
	regCommon
	legacyObjs []T
}

// ToLegacyObjects returns the legacy properties of the SimsRegistryAdapter as a slice of type T.
func (l *AbstractRegistry[T]) ToLegacyObjects() []T {
	return l.legacyObjs
}

// WeightedOperationRegistryAdapter is an implementation of the Registry interface that provides adapters to use the new message factories
// with the legacy simulation system
type WeightedOperationRegistryAdapter struct {
	AbstractRegistry[simtypes.WeightedOperation]
}

// NewSimsMsgRegistryAdapter creates a new instance of SimsRegistryAdapter for WeightedOperation types.
func NewSimsMsgRegistryAdapter(
	reporter SimulationReporter,
	ak AccountSourceX,
	bk BalanceSource,
	txConfig client.TxConfig,
	logger log.Logger,
) *WeightedOperationRegistryAdapter {
	return &WeightedOperationRegistryAdapter{
		AbstractRegistry: AbstractRegistry[simtypes.WeightedOperation]{
			regCommon: regCommon{
				reporter:     reporter,
				ak:           ak,
				bk:           bk,
				txConfig:     txConfig,
				addressCodec: txConfig.SigningContext().AddressCodec(),
				logger:       logger,
			},
		},
	}
}

// Add adds a new weighted operation to the collection
func (l *WeightedOperationRegistryAdapter) Add(weight uint32, fx SimMsgFactoryX) {
	if fx == nil {
		panic("message factory must not be nil")
	}
	if weight == 0 {
		return
	}
	obj := simulation.NewWeightedOperation(int(weight), legacyOperationAdapter(l.regCommon, fx))
	l.legacyObjs = append(l.legacyObjs, obj)
}

type hasFutureOpsRegistry interface {
	SetFutureOpsRegistry(FutureOpsRegistry)
}

func legacyOperationAdapter(l regCommon, fx SimMsgFactoryX) simtypes.Operation {
	return func(
		r *rand.Rand, app AppEntrypoint, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		xCtx, done := context.WithCancel(ctx)
		ctx = sdk.UnwrapSDKContext(xCtx)
		testData := l.newChainDataSource(ctx, r, accs...)
		reporter := l.reporter.WithScope(fx.MsgType(), SkipHookFn(func(args ...any) { done() }))
		fOpsReg := newFutureOpsRegistry(l)
		if fx, ok := fx.(hasFutureOpsRegistry); ok {
			fx.SetFutureOpsRegistry(fOpsReg)
		}
		from, msg := runWithFailFast(ctx, testData, reporter, fx.Create())
		futOps := fOpsReg.legacyObjs
		weightedOpsResult := DeliverSimsMsg(ctx, reporter, app, r, l.txConfig, l.ak, chainID, msg, fx.DeliveryResultHandler(), from...)
		err := reporter.Close()
		return weightedOpsResult, futOps, err
	}
}

func newFutureOpsRegistry(l regCommon) *FutureOperationRegistryAdapter {
	return &FutureOperationRegistryAdapter{regCommon: l}
}

type FutureOperationRegistryAdapter AbstractRegistry[simtypes.FutureOperation]

func (l *FutureOperationRegistryAdapter) Add(blockTime time.Time, fx SimMsgFactoryX) {
	if fx == nil {
		panic("message factory must not be nil")
	}
	if blockTime.IsZero() {
		return
	}
	obj := simtypes.FutureOperation{
		BlockTime: blockTime,
		Op:        legacyOperationAdapter(l.regCommon, fx),
	}
	l.legacyObjs = append(l.legacyObjs, obj)
}

type SimsProposalRegistryAdapter struct {
	AbstractRegistry[simtypes.WeightedProposalMsg]
}

// NewSimsProposalRegistryAdapter creates a new instance of SimsRegistryAdapter for WeightedProposalMsg types.
func NewSimsProposalRegistryAdapter(
	reporter SimulationReporter,
	ak AccountSourceX,
	bk BalanceSource,
	addrCodec address.Codec,
	logger log.Logger,
) *SimsProposalRegistryAdapter {
	return &SimsProposalRegistryAdapter{
		AbstractRegistry: AbstractRegistry[simtypes.WeightedProposalMsg]{
			regCommon: regCommon{
				reporter:     reporter,
				ak:           ak,
				bk:           bk,
				addressCodec: addrCodec,
				logger:       logger,
			},
		},
	}
}

func (l *SimsProposalRegistryAdapter) Add(weight uint32, fx SimMsgFactoryX) {
	if fx == nil {
		panic("message factory must not be nil")
	}
	if weight == 0 {
		return
	}
	l.legacyObjs = append(l.legacyObjs, legacyProposalMsgAdapter(l.regCommon, weight, fx))
}

// legacyProposalMsgAdapter adapter to convert the new msg factory into the weighted proposal message type
func legacyProposalMsgAdapter(l regCommon, weight uint32, fx SimMsgFactoryX) simtypes.WeightedProposalMsg {
	msgAdapter := func(ctx context.Context, r *rand.Rand, accs []simtypes.Account, cdc address.Codec) (sdk.Msg, error) {
		xCtx, done := context.WithCancel(ctx)
		testData := l.newChainDataSource(xCtx, r, accs...)
		reporter := l.reporter.WithScope(fx.MsgType(), SkipHookFn(func(args ...any) { done() }))
		_, msg := runWithFailFast(xCtx, testData, reporter, fx.Create())
		return msg, reporter.Close()
	}
	return simulation.NewWeightedProposalMsgX("", int(weight), msgAdapter)
}

type tuple struct {
	signer []SimAccount
	msg    sdk.Msg
}

// runWithFailFast runs the factory method on a separate goroutine to abort early when the context is canceled via reporter skip
func runWithFailFast(
	ctx context.Context,
	data *ChainDataSource,
	reporter SimulationReporter,
	f FactoryMethod,
) (signer []SimAccount, msg sdk.Msg) {
	r := make(chan tuple)
	go func() {
		defer recoverPanicForSkipped(reporter, r)
		signer, msg := f(ctx, data, reporter)
		r <- tuple{signer: signer, msg: msg}
	}()
	select {
	case t, ok := <-r:
		if !ok {
			return nil, nil
		}
		return t.signer, t.msg
	case <-ctx.Done():
		reporter.Skip("context closed")
		return nil, nil
	}
}

func recoverPanicForSkipped(reporter SimulationReporter, resultChan chan tuple) {
	if r := recover(); r != nil {
		if !reporter.IsSkipped() {
			panic(r)
		}
		close(resultChan)
	}
}
