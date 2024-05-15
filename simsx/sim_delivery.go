package simsx

import (
	"context"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type AppEntrypoint interface {
	SimDeliver(_txEncoder sdk.TxEncoder, tx sdk.Tx) (sdk.GasInfo, *sdk.Result, error)
}
type AccountSource interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
}

func DeliverSimsMsg(
	reporter SimulationReporter,
	r *rand.Rand,
	app AppEntrypoint,
	txGen client.TxConfig,
	ak AccountSource,
	msg sdk.Msg,
	ctx sdk.Context,
	chainID string,
	senders ...SimAccount,
) simtypes.OperationMsg {
	if reporter.IsSkipped() {
		return reporter.ToLegacyOperationMsg()
	}
	accountNumbers := make([]uint64, len(senders))
	sequenceNumbers := make([]uint64, len(senders))
	for i := 0; i < len(senders); i++ {
		acc := ak.GetAccount(ctx, senders[i].Address)
		accountNumbers[i] = acc.GetAccountNumber()
		sequenceNumbers[i] = acc.GetSequence()
	}

	fees := senders[0].LiquidBalance().RandFees()
	tx, err := sims.GenSignedMockTx(
		r,
		txGen,
		[]sdk.Msg{msg},
		fees,
		sims.DefaultGenTxGas,
		chainID,
		accountNumbers,
		sequenceNumbers,
		Collect(senders, func(a SimAccount) cryptotypes.PrivKey { return a.PrivKey })...,
	)
	if err != nil {
		reporter.Fail(err, "encoding TX")
		return reporter.ToLegacyOperationMsg()
	}
	_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
	if err != nil {
		reporter.Fail(err, "delivering tx")
		return reporter.ToLegacyOperationMsg()
	}
	reporter.Success(msg)
	return reporter.ToLegacyOperationMsg()
}
