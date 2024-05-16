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

type SimMsgFactoryX interface {
	MsgType() sdk.Msg
	Create() FactoryMethod
}
type Registry interface {
	Add(weight uint32, f SimMsgFactoryX)
	ToLegacyWeightedOperations() simulation.WeightedOperations
}

var _ Registry = &SimsRegistryAdapter{}

type xAccountSource interface {
	AccountSource
	ModuleAccountSource
	AddressCodec() address.Codec
}
type SimsRegistryAdapter struct {
	reporter  SimulationReporter
	legacyOps simulation.WeightedOperations
	ak        xAccountSource
	bk        BalanceSource
	txConfig  client.TxConfig
}

func NewSimsRegistryAdapter(reporter SimulationReporter, ak xAccountSource, bk BalanceSource, txConfig client.TxConfig) *SimsRegistryAdapter {
	return &SimsRegistryAdapter{
		reporter: reporter,
		ak:       ak,
		bk:       bk,
		txConfig: txConfig,
	}
}

func (l *SimsRegistryAdapter) Add(weight uint32, f SimMsgFactoryX) {
	if f == nil {
		panic("message factory is nil")
	}
	opAdapter := l.newLegacyOperationAdapter(l.reporter, f.MsgType(), f.Create())
	l.legacyOps = append(l.legacyOps, simulation.NewWeightedOperation(int(weight), opAdapter))
}

func (l SimsRegistryAdapter) newLegacyOperationAdapter(rootReporter SimulationReporter, example sdk.Msg, f FactoryMethod) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		testData := NewChainDataSource(r, l.ak, NewBalanceSource(ctx, l.bk), l.ak.AddressCodec(), accs...)
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
		), nil, reporter.ExecutionResult().Error
	}
}

func (l SimsRegistryAdapter) ToLegacyWeightedOperations() simulation.WeightedOperations {
	return l.legacyOps
}
