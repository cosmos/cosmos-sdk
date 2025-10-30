package txnrunner

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"

	storetypes "cosmossdk.io/store/types"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.TxRunner = DefaultRunner{}

func NewDefaultRunner(txDecoder sdk.TxDecoder, tracer log.Tracer) *DefaultRunner {
	return &DefaultRunner{
		txDecoder: txDecoder,
		tracer:    tracer,
	}
}

// DefaultRunner is the default TxnRunner implementation which executes the transactions in a block sequentially.
type DefaultRunner struct {
	txDecoder sdk.TxDecoder
	tracer    log.Tracer
}

func (d DefaultRunner) Run(ctx context.Context, _ storetypes.MultiStore, txs [][]byte, deliverTx sdk.DeliverTxFunc) ([]*abci.ExecTxResult, error) {
	// Fallback to the default execution logic
	txResults := make([]*abci.ExecTxResult, 0, len(txs))
	for i, rawTx := range txs {
		var response *abci.ExecTxResult

		if _, err := d.txDecoder(rawTx); err == nil {
			response = deliverTx(rawTx, nil, i, nil, d.tracer)
		} else {
			// In the case where a transaction included in a block proposal is malformed,
			// we still want to return a default response to comet. This is because comet
			// expects a response for each transaction included in a block proposal.
			response = sdkerrors.ResponseExecTxResultWithEvents(
				sdkerrors.ErrTxDecode,
				0,
				0,
				nil,
				false,
			)
		}

		// check after every tx if we should abort
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// continue
		}

		txResults = append(txResults, response)
	}
	return txResults, nil
}
