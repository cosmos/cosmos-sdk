package simsx

import (
	"cmp"
	"context"
	"iter"
	"maps"
	"math/rand"
	"slices"
	"strings"
	"time"

	"cosmossdk.io/core/address"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

type (
	// Registry is an abstract entry point to register message factories with weights
	Registry interface {
		Add(weight uint32, f SimMsgFactoryX)
	}
	// FutureOpsRegistry register message factories for future blocks
	FutureOpsRegistry interface {
		Add(blockTime time.Time, f SimMsgFactoryX)
	}

	// AccountSourceX Account and Module account
	AccountSourceX interface {
		AccountSource
		ModuleAccountSource
	}
)

// WeightedProposalMsgIter iterator for weighted gov proposal payload messages
type WeightedProposalMsgIter = iter.Seq2[uint32, FactoryMethod]

var _ Registry = &WeightedOperationRegistryAdapter{}

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
	items []T
}

type operation func(
	r *rand.Rand, app AppEntrypoint, ctx sdk.Context,
	accs []simtypes.Account, chainID string,
) (simtypes.OperationMsg, []simtypes.FutureOperation, error)

// make api compatible to simtypes.Operation
func (o operation) toLegacyOp() simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accounts []simtypes.Account, chainID string) (OperationMsg simtypes.OperationMsg, futureOps []simtypes.FutureOperation, err error) {
		return o(r, app, ctx, accounts, chainID)
	}
}

var _ simtypes.WeightedOperation = weightedOperation{}

type weightedOperation struct {
	weight uint32
	op     operation
}

func (w weightedOperation) Weight() int {
	return int(w.weight)
}

func (w weightedOperation) Op() simtypes.Operation {
	return w.op.toLegacyOp()
}

// WeightedOperationRegistryAdapter is an implementation of the Registry interface that provides adapters to use the new message factories
// with the legacy simulation system
type WeightedOperationRegistryAdapter struct {
	AbstractRegistry[weightedOperation]
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
		AbstractRegistry: AbstractRegistry[weightedOperation]{
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
	obj := weightedOperation{weight: weight, op: legacyOperationAdapter(l.regCommon, fx)}
	l.items = append(l.items, obj)
}

type HasFutureOpsRegistry interface {
	SetFutureOpsRegistry(FutureOpsRegistry)
}

// msg factory to legacy Operation type
func legacyOperationAdapter(l regCommon, fx SimMsgFactoryX) operation {
	return func(
		r *rand.Rand, app AppEntrypoint, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		xCtx, done := context.WithCancel(ctx)
		ctx = sdk.UnwrapSDKContext(xCtx)
		testData := l.newChainDataSource(ctx, r, accs...)
		reporter := l.reporter.WithScope(fx.MsgType(), SkipHookFn(func(args ...any) { done() }))
		fOpsReg := NewFutureOpsRegistry(l)
		if fx, ok := fx.(HasFutureOpsRegistry); ok {
			fx.SetFutureOpsRegistry(fOpsReg)
		}
		from, msg := SafeRunFactoryMethod(ctx, testData, reporter, fx.Create())
		futOps := fOpsReg.items
		weightedOpsResult := DeliverSimsMsg(ctx, reporter, app, r, l.txConfig, l.ak, chainID, msg, fx.DeliveryResultHandler(), from...)
		err := reporter.Close()
		return weightedOpsResult, futOps, err
	}
}

func NewFutureOpsRegistry(l regCommon) *FutureOperationRegistryAdapter {
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
		Op:        legacyOperationAdapter(l.regCommon, fx).toLegacyOp(),
	}
	l.items = append(l.items, obj)
}

var _ Registry = &UniqueTypeRegistry{}

type UniqueTypeRegistry map[string]WeightedFactory

func NewUniqueTypeRegistry() UniqueTypeRegistry {
	return make(UniqueTypeRegistry)
}

func (s UniqueTypeRegistry) Add(weight uint32, f SimMsgFactoryX) {
	if weight == 0 {
		return
	}
	if f == nil {
		panic("message factory must not be nil")
	}
	msgType := f.MsgType()
	msgTypeURL := sdk.MsgTypeURL(msgType)
	if _, exists := s[msgTypeURL]; exists {
		panic("type is already registered: " + msgTypeURL)
	}
	s[msgTypeURL] = WeightedFactory{Weight: weight, Factory: f}
}

// Iterator returns an iterator function for a Go for loop sorted by weight desc.
func (s UniqueTypeRegistry) Iterator() WeightedProposalMsgIter {
	sortedWeightedFactory := slices.SortedFunc(maps.Values(s), func(a, b WeightedFactory) int {
		return a.Compare(b)
	})

	return func(yield func(uint32, FactoryMethod) bool) {
		for _, v := range sortedWeightedFactory {
			if !yield(v.Weight, v.Factory.Create()) {
				return
			}
		}
	}
}

var _ Registry = &UnorderedRegistry{}

// UnorderedRegistry represents a collection of WeightedFactory elements without guaranteed order.
// It is used to maintain factories coupled with their associated weights for simulation purposes.
type UnorderedRegistry []WeightedFactory

func NewUnorderedRegistry() *UnorderedRegistry {
	r := make(UnorderedRegistry, 0)
	return &r
}

// Add appends a new WeightedFactory with the provided weight and factory to the UnorderedRegistry.
func (x *UnorderedRegistry) Add(weight uint32, f SimMsgFactoryX) {
	if weight == 0 {
		return
	}
	if f == nil {
		panic("message factory must not be nil")
	}
	*x = append(*x, WeightedFactory{Weight: weight, Factory: f})
}

// Elements returns all collected elements
func (x *UnorderedRegistry) Elements() []WeightedFactory {
	return *x
}

// WeightedFactory is a Weight Factory tuple
type WeightedFactory struct {
	Weight  uint32
	Factory SimMsgFactoryX
}

// Compare compares the WeightedFactory f with another WeightedFactory b.
// The comparison is done by comparing the weight of f with the weight of b.
// If the weight of f is greater than the weight of b, it returns 1.
// If the weight of f is less than the weight of b, it returns -1.
// If the weights are equal, it compares the TypeURL of the factories using strings.Compare.
// Returns an integer indicating the comparison result.
func (f WeightedFactory) Compare(b WeightedFactory) int {
	switch {
	case f.Weight > b.Weight:
		return 1
	case f.Weight < b.Weight:
		return -1
	default:
		return strings.Compare(sdk.MsgTypeURL(f.Factory.MsgType()), sdk.MsgTypeURL(b.Factory.MsgType()))
	}
}

// WeightedFactoryMethod is a data tuple used for registering legacy proposal operations
type WeightedFactoryMethod struct {
	Weight  uint32
	Factory FactoryMethod
}

type WeightedFactoryMethods []WeightedFactoryMethod

// NewWeightedFactoryMethods constructor
func NewWeightedFactoryMethods() WeightedFactoryMethods {
	return make(WeightedFactoryMethods, 0)
}

// Add adds a new WeightedFactoryMethod to the WeightedFactoryMethods slice.
// If weight is zero or f is nil, it returns without making any changes.
func (s *WeightedFactoryMethods) Add(weight uint32, f FactoryMethod) {
	if weight == 0 {
		return
	}
	if f == nil {
		panic("message factory must not be nil")
	}
	*s = append(*s, WeightedFactoryMethod{Weight: weight, Factory: f})
}

// Iterator returns an iterator function for a Go for loop sorted by weight desc.
func (s WeightedFactoryMethods) Iterator() WeightedProposalMsgIter {
	slices.SortFunc(s, func(e, e2 WeightedFactoryMethod) int {
		return cmp.Compare(e.Weight, e2.Weight)
	})
	return func(yield func(uint32, FactoryMethod) bool) {
		for _, v := range s {
			if !yield(v.Weight, v.Factory) {
				return
			}
		}
	}
}

// legacy operation to Msg factory type
func legacyToMsgFactoryAdapter(fn simtypes.MsgSimulatorFn) FactoryMethod {
	return func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter) (signer []SimAccount, msg sdk.Msg) {
		return []SimAccount{}, fn(testData.r, sdk.UnwrapSDKContext(ctx), testData.AllAccounts())
	}
}

// AppendIterators takes multiple WeightedProposalMsgIter and returns a single iterator that sequentially yields items after each one.
func AppendIterators(iterators ...WeightedProposalMsgIter) WeightedProposalMsgIter {
	return func(yield func(uint32, FactoryMethod) bool) {
		for _, it := range iterators {
			it(yield)
		}
	}
}
