package baseapp

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// TxExecutor function type for implementing custom execution logic, such as block-stm
type TxExecutor interface {
	run(txs [][]byte) ([]*abci.ExecTxResult, error)
}

type DefaultExecutor struct {
	ctx       context.Context
	txDecoder sdk.TxDecoder
	deliverTx func(tx []byte) *abci.ExecTxResult
}

func (d DefaultExecutor) run(txs [][]byte) ([]*abci.ExecTxResult, error) {
	// Fallback to the default execution logic
	txResults := make([]*abci.ExecTxResult, 0, len(txs))
	for _, rawTx := range txs {
		var response *abci.ExecTxResult

		if _, err := d.txDecoder(rawTx); err == nil {
			response = d.deliverTx(rawTx)
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
		case <-d.ctx.Done():
			return nil, d.ctx.Err()
		default:
			// continue
		}

		txResults = append(txResults, response)
	}
	return txResults, nil
}
