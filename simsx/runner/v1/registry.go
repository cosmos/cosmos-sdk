package v1

import (
	"cmp"
	"context"
	"github.com/cosmos/cosmos-sdk/simsx/common"
	common2 "github.com/cosmos/cosmos-sdk/simsx/runner/common"
	"math/rand"
	"slices"
	"time"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/log"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

type (
	Registry                = common.Registry
	FutureOpsRegistry       = common.FutureOpsRegistry
	HasFutureOpsRegistry    = common.HasFutureOpsRegistry
	WeightedProposalMsgIter = common.WeightedProposalMsgIter
)

var _ Registry = &WeightedOperationRegistryAdapter{}

// common types for abstract registry without generics
type regCommon struct {
	reporter     common.SimulationReporterRuntime
	ak           common.AccountSourceX
	bk           common.BalanceSource
	addressCodec address.Codec
	txConfig     client.TxConfig
	logger       log.Logger
}

func (c regCommon) newChainDataSource(ctx context.Context, r *rand.Rand, accs ...simtypes.Account) *common.ChainDataSource {
	return common.NewChainDataSource(ctx, r, c.ak, c.bk, c.addressCodec, accs...)
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
	reporter common.SimulationReporterRuntime,
	ak common.AccountSourceX,
	bk common.BalanceSource,
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
func (l *WeightedOperationRegistryAdapter) Add(weight uint32, fx common.SimMsgFactoryX) {
	if fx == nil {
		panic("message factory must not be nil")
	}
	if weight == 0 {
		return
	}
	obj := simulation.NewWeightedOperation(int(weight), legacyOperationAdapter(l.regCommon, fx))
	l.legacyObjs = append(l.legacyObjs, obj)
}

// msg factory to legacy Operation type
func legacyOperationAdapter(l regCommon, fx common.SimMsgFactoryX) simtypes.Operation {
	return func(
		r *rand.Rand, app AppEntrypoint, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		xCtx, done := context.WithCancel(ctx)
		ctx = sdk.UnwrapSDKContext(xCtx)
		testData := l.newChainDataSource(ctx, r, accs...)
		reporter := l.reporter.WithScope(fx.MsgType(), common.SkipHookFn(func(args ...any) { done() }))
		fOpsReg := NewFutureOpsRegistry(l)
		if fx, ok := fx.(HasFutureOpsRegistry); ok {
			fx.SetFutureOpsRegistry(fOpsReg)
		}
		from, msg := common2.SafeRunFactoryMethod(ctx, testData, reporter, fx.Create())
		futOps := fOpsReg.legacyObjs
		weightedOpsResult := DeliverSimsMsg(ctx, reporter, app, r, l.txConfig, l.ak, chainID, msg, fx.DeliveryResultHandler(), from...)
		err := reporter.Close()
		return weightedOpsResult, futOps, err
	}
}

func NewFutureOpsRegistry(l regCommon) *FutureOperationRegistryAdapter {
	return &FutureOperationRegistryAdapter{regCommon: l}
}

type FutureOperationRegistryAdapter AbstractRegistry[simtypes.FutureOperation]

func (l *FutureOperationRegistryAdapter) Add(blockTime time.Time, fx common.SimMsgFactoryX) {
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

// WeightedFactoryMethod is a data tuple used for registering legacy proposal operations
type WeightedFactoryMethod struct {
	Weight  uint32
	Factory common.FactoryMethod
}

type WeightedFactoryMethods []WeightedFactoryMethod

// NewWeightedFactoryMethods constructor
func NewWeightedFactoryMethods() WeightedFactoryMethods {
	return make(WeightedFactoryMethods, 0)
}

// Add adds a new WeightedFactoryMethod to the WeightedFactoryMethods slice.
// If weight is zero or f is nil, it returns without making any changes.
func (s *WeightedFactoryMethods) Add(weight uint32, f common.FactoryMethod) {
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
	return func(yield func(uint32, common.FactoryMethod) bool) {
		for _, v := range s {
			if !yield(v.Weight, v.Factory) {
				return
			}
		}
	}
}

// legacy operation to Msg factory type
func legacyToMsgFactoryAdapter(fn simtypes.MsgSimulatorFnX) common.FactoryMethod {
	return func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) (signer []common.SimAccount, msg sdk.Msg) {
		msg, err := fn(ctx, testData.Rand().Rand, testData.AllAccounts(), testData.AddressCodec())
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		return []common.SimAccount{}, msg
	}
}

// AppendIterators takes multiple WeightedProposalMsgIter and returns a single iterator that sequentially yields items after each one.
func AppendIterators(iterators ...WeightedProposalMsgIter) WeightedProposalMsgIter {
	return func(yield func(uint32, common.FactoryMethod) bool) {
		for _, it := range iterators {
			it(yield)
		}
	}
}
