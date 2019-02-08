package context

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BroadcastTx broadcasts a transactions either synchronously or asynchronously
// based on the context parameters. The result of the broadcast is parsed into
// an intermediate structure which is logged if the context has a logger
// defined.
func (ctx CLIContext) BroadcastTx(txBytes []byte) (res sdk.TxResponse, err error) {
	if ctx.Async {
		if res, err = ctx.BroadcastTxAsync(txBytes); err != nil {
			return
		}
		return
	}

	if res, err = ctx.BroadcastTxAndAwaitCommit(txBytes); err != nil {
		return
	}

	return
}

// BroadcastTxAndAwaitCommit broadcasts transaction bytes to a Tendermint node
// and waits for a commit.
func (ctx CLIContext) BroadcastTxAndAwaitCommit(tx []byte) (sdk.TxResponse, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return sdk.TxResponse{}, err
	}

	res, err := node.BroadcastTxCommit(tx)
	if err != nil {
		return sdk.NewResponseFormatBroadcastTxCommit(res), err
	}

	if !res.CheckTx.IsOK() {
		return sdk.NewResponseFormatBroadcastTxCommit(res), fmt.Errorf(res.CheckTx.Log)
	}

	if !res.DeliverTx.IsOK() {
		return sdk.NewResponseFormatBroadcastTxCommit(res), fmt.Errorf(res.DeliverTx.Log)
	}

	return sdk.NewResponseFormatBroadcastTxCommit(res), err
}

// BroadcastTxSync broadcasts transaction bytes to a Tendermint node synchronously.
func (ctx CLIContext) BroadcastTxSync(tx []byte) (sdk.TxResponse, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return sdk.TxResponse{}, err
	}

	res, err := node.BroadcastTxSync(tx)
	if err != nil {
		return sdk.NewResponseFormatBroadcastTx(res), err
	}

	return sdk.NewResponseFormatBroadcastTx(res), err
}

// BroadcastTxAsync broadcasts transaction bytes to a Tendermint node asynchronously.
func (ctx CLIContext) BroadcastTxAsync(tx []byte) (sdk.TxResponse, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return sdk.TxResponse{}, err
	}

	res, err := node.BroadcastTxAsync(tx)
	if err != nil {
		return sdk.NewResponseFormatBroadcastTx(res), err
	}

	return sdk.NewResponseFormatBroadcastTx(res), err
}
