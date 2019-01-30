package context

import (
	"fmt"
	"io"

	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// TODO: This should get deleted eventually, and perhaps
// ctypes.ResultBroadcastTx be stripped of unused fields, and
// ctypes.ResultBroadcastTxCommit returned for tendermint RPC BroadcastTxSync.
//
// The motivation is that we want a unified type to return, and the better
// option is the one that can hold CheckTx/DeliverTx responses optionally.
func resultBroadcastTxToCommit(res *ctypes.ResultBroadcastTx) *ctypes.ResultBroadcastTxCommit {
	return &ctypes.ResultBroadcastTxCommit{
		Hash: res.Hash,
		// NOTE: other fields are unused for async.
	}
}

// BroadcastTx broadcasts a transactions either synchronously or asynchronously
// based on the context parameters. The result of the broadcast is parsed into
// an intermediate structure which is logged if the context has a logger
// defined.
func (ctx CLIContext) BroadcastTx(txBytes []byte) (*ctypes.ResultBroadcastTxCommit, error) {
	if ctx.Async {
		res, err := ctx.broadcastTxAsync(txBytes)
		if err != nil {
			return nil, err
		}

		resCommit := resultBroadcastTxToCommit(res)
		return resCommit, err
	}

	return ctx.broadcastTxCommit(txBytes)
}

// BroadcastTxAndAwaitCommit broadcasts transaction bytes to a Tendermint node
// and waits for a commit.
func (ctx CLIContext) BroadcastTxAndAwaitCommit(tx []byte) (*ctypes.ResultBroadcastTxCommit, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return nil, err
	}

	res, err := node.BroadcastTxCommit(tx)
	if err != nil {
		return res, err
	}

	if !res.CheckTx.IsOK() {
		return res, errors.Errorf(res.CheckTx.Log)
	}

	if !res.DeliverTx.IsOK() {
		return res, errors.Errorf(res.DeliverTx.Log)
	}

	return res, err
}

// BroadcastTxSync broadcasts transaction bytes to a Tendermint node
// synchronously.
func (ctx CLIContext) BroadcastTxSync(tx []byte) (*ctypes.ResultBroadcastTx, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return nil, err
	}

	res, err := node.BroadcastTxSync(tx)
	if err != nil {
		return res, err
	}

	return res, err
}

// BroadcastTxAsync broadcasts transaction bytes to a Tendermint node
// asynchronously.
func (ctx CLIContext) BroadcastTxAsync(tx []byte) (*ctypes.ResultBroadcastTx, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return nil, err
	}

	res, err := node.BroadcastTxAsync(tx)
	if err != nil {
		return res, err
	}

	return res, err
}

func (ctx CLIContext) broadcastTxAsync(txBytes []byte) (*ctypes.ResultBroadcastTx, error) {
	res, err := ctx.BroadcastTxAsync(txBytes)
	if err != nil {
		return res, err
	}

	if ctx.OutputIsSet() && ctx.OutputIsJSON() {
		bz, err := ctx.Codec.MarshalJSON(struct {
			T string `json:"tx_hash"`
		}{T: res.Hash.String()})

		if err != nil {
			return res, err
		}

		ctx.Output.Write(bz)
		io.WriteString(ctx.Output, "\n")
	}

	if ctx.OutputIsSet() {
		io.WriteString(ctx.Output, fmt.Sprintf("async tx with hash %s sent\n", res.Hash))
	}

	return res, nil
}

func (ctx CLIContext) broadcastTxCommit(txBytes []byte) (*ctypes.ResultBroadcastTxCommit, error) {
	res, err := ctx.BroadcastTxAndAwaitCommit(txBytes)
	if err != nil {
		return res, err
	}

	if ctx.OutputIsJSON() && ctx.OutputIsSet() {
		// Since JSON is intended for automated scripts, always include response in
		// JSON mode.

		bz, err := ctx.Codec.MarshalJSON(sdk.NewStringResponseDeliverTx(res))
		if err != nil {
			return res, err
		}

		ctx.Output.Write(bz)
		io.WriteString(ctx.Output, "\n")

		return res, nil
	}

	if ctx.OutputIsSet() {
		resStr := fmt.Sprintf("Committed tx with hash %s at block %d\n", res.Height, res.Hash.String())

		if ctx.PrintResponse {
			resStr += fmt.Sprintf("Response: %+#v\n", res.DeliverTx)
		}

		io.WriteString(ctx.Output, resStr)
	}

	return res, nil
}
