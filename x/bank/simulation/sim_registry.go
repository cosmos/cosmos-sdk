package simulation

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

type SimMsgFactory func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter) ([]SimAccount, sdk.Msg)

type Registry interface {
	Add(weight uint32, example sdk.Msg, f SimMsgFactory)
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

func NewSimsRegistryAdapter(
	reporter SimulationReporter,
	ak xAccountSource,
	bk BalanceSource,
	txConfig client.TxConfig,
) *SimsRegistryAdapter {
	return &SimsRegistryAdapter{
		reporter: reporter,
		ak:       ak,
		bk:       bk,
		txConfig: txConfig,
	}
}

func (l *SimsRegistryAdapter) Add(weight uint32, example sdk.Msg, f SimMsgFactory) {
	if example == nil {
		panic("example is nil")
	}
	if f == nil {
		panic("message factory is nil")
	}
	opAdapter := l.newLegacyOperationAdapter(l.reporter, example, f)
	l.legacyOps = append(l.legacyOps, simulation.NewWeightedOperation(int(weight), opAdapter))
}

func (l SimsRegistryAdapter) newLegacyOperationAdapter(reporter SimulationReporter, example sdk.Msg, f SimMsgFactory) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		testData := NewChainDataSource(r, l.ak, NewBalanceSource(ctx, l.bk), l.ak.AddressCodec(), accs...)
		reporter = reporter.WithScope(example)

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
