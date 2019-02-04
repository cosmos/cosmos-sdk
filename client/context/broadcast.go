package context

import (
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BroadcastTx broadcasts a transactions either synchronously or asynchronously
// based on the context parameters. The result of the broadcast is parsed into
// an intermediate structure which is logged if the context has a logger
// defined.
func (ctx CLIContext) BroadcastTx(txBytes []byte) (res sdk.ResponseDeliverTx, err error) {
	if ctx.Async {
		if res, err = ctx.BroadcastTxAsync(txBytes); err != nil {
			return
		}
		return
	}

	if res, err = ctx.BroadcastTxAndAwaitCommit(txBytes); err != nil {
		return
	}

	ctx.PrintOutput(res)
	return
}

// BroadcastTxAndAwaitCommit broadcasts transaction bytes to a Tendermint node
// and waits for a commit.
func (ctx CLIContext) BroadcastTxAndAwaitCommit(tx []byte) (sdk.ResponseDeliverTx, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return sdk.ResponseDeliverTx{}, err
	}

	res, err := node.BroadcastTxCommit(tx)
	if err != nil {
		return sdk.NewResponseDeliverTxCommit(res), err
	}

	if !res.CheckTx.IsOK() {
		return sdk.NewResponseDeliverTxCommit(res), errors.Errorf(res.CheckTx.Log)
	}

	if !res.DeliverTx.IsOK() {
		return sdk.NewResponseDeliverTxCommit(res), errors.Errorf(res.DeliverTx.Log)
	}

	return sdk.NewResponseDeliverTxCommit(res), err
}

// BroadcastTxSync broadcasts transaction bytes to a Tendermint node synchronously.
func (ctx CLIContext) BroadcastTxSync(tx []byte) (sdk.ResponseDeliverTx, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return sdk.ResponseDeliverTx{}, err
	}

	res, err := node.BroadcastTxSync(tx)
	if err != nil {
		return sdk.NewResponseDeliverTx(res), err
	}

	return sdk.NewResponseDeliverTx(res), err
}

// BroadcastTxAsync broadcasts transaction bytes to a Tendermint node asynchronously.
func (ctx CLIContext) BroadcastTxAsync(tx []byte) (sdk.ResponseDeliverTx, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return sdk.ResponseDeliverTx{}, err
	}

	res, err := node.BroadcastTxAsync(tx)
	if err != nil {
		return sdk.NewResponseDeliverTx(res), err
	}

	return sdk.NewResponseDeliverTx(res), err
}
