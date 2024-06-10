package simsx

import (
	"context"
	"math/rand"

	"cosmossdk.io/core/address"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

type FactoryMethod func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter) (signer []SimAccount, msg sdk.Msg)

var _ SimMsgFactoryX = SimMsgFactoryFn[sdk.Msg](nil)

type SimMsgFactoryFn[T sdk.Msg] FactoryMethod

func (f SimMsgFactoryFn[T]) MsgType() sdk.Msg {
	var x T
	return x
}

func (f SimMsgFactoryFn[T]) Create() FactoryMethod {
	return FactoryMethod(f)
}

func (f SimMsgFactoryFn[T]) Cast(msg sdk.Msg) T {
	return msg.(T)
}

type SimMsgFactoryX interface {
	MsgType() sdk.Msg
	Create() FactoryMethod
}
type Registry interface {
	Add(weight uint32, f SimMsgFactoryX)
	ToLegacyWeightedOperations() simulation.WeightedOperations
}

var _ Registry = &SimsRegistryAdapter{}

type AccountSourceX interface {
	AccountSource
	ModuleAccountSource
}
type SimsRegistryAdapter struct {
	reporter     SimulationReporter
	legacyOps    simulation.WeightedOperations
	ak           AccountSourceX
	bk           BalanceSource
	txConfig     client.TxConfig
	addressCodec address.Codec
}

func NewSimsRegistryAdapter(reporter SimulationReporter, ak AccountSourceX, bk BalanceSource, txConfig client.TxConfig) *SimsRegistryAdapter {
	return &SimsRegistryAdapter{
		reporter:     reporter,
		ak:           ak,
		bk:           bk,
		txConfig:     txConfig,
		addressCodec: txConfig.SigningContext().AddressCodec(),
	}
}

func (l *SimsRegistryAdapter) Add(weight uint32, f SimMsgFactoryX) {
	if f == nil {
		panic("message factory is nil")
	}
	opAdapter := l.legacyOperationAdapter(l.reporter, f.MsgType(), f.Create())
	l.legacyOps = append(l.legacyOps, simulation.NewWeightedOperation(int(weight), opAdapter))
}

// adapter to convert the new msg factory into the weighted operations
func (l SimsRegistryAdapter) legacyOperationAdapter(rootReporter SimulationReporter, example sdk.Msg, f FactoryMethod) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		testData := NewChainDataSource(ctx, r, l.ak, l.bk, l.addressCodec, accs...)
		reporter := rootReporter.WithScope(example)

		from, msg := f(ctx, testData, reporter)
		return DeliverSimsMsg(
			reporter,
			r,
			app,
			l.txConfig,
			l.ak,
			msg,
			ctx,
			chainID,
			from...,
		), nil, reporter.Close()
	}
}

func (l SimsRegistryAdapter) ToLegacyWeightedOperations() simulation.WeightedOperations {
	return l.legacyOps
}
