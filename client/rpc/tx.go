package rpc

import (
	"context"
	"encoding/hex"
	"strings"
	"time"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

func newTxResponseCheckTx(res *coretypes.ResultBroadcastTxCommit) *sdk.TxResponse {
	if res == nil {
		return nil
	}

	var txHash string
	if res.Hash != nil {
		txHash = res.Hash.String()
	}

	parsedLogs, _ := sdk.ParseABCILogs(res.CheckTx.Log)

	return &sdk.TxResponse{
		Height:    res.Height,
		TxHash:    txHash,
		Codespace: res.CheckTx.Codespace,
		Code:      res.CheckTx.Code,
		Data:      strings.ToUpper(hex.EncodeToString(res.CheckTx.Data)),
		RawLog:    res.CheckTx.Log,
		Logs:      parsedLogs,
		Info:      res.CheckTx.Info,
		GasWanted: res.CheckTx.GasWanted,
		GasUsed:   res.CheckTx.GasUsed,
		Events:    res.CheckTx.Events,
	}
}

func newTxResponseDeliverTx(res *coretypes.ResultBroadcastTxCommit) *sdk.TxResponse {
	if res == nil {
		return nil
	}

	var txHash string
	if res.Hash != nil {
		txHash = res.Hash.String()
	}

	parsedLogs, _ := sdk.ParseABCILogs(res.DeliverTx.Log)

	return &sdk.TxResponse{
		Height:    res.Height,
		TxHash:    txHash,
		Codespace: res.DeliverTx.Codespace,
		Code:      res.DeliverTx.Code,
		Data:      strings.ToUpper(hex.EncodeToString(res.DeliverTx.Data)),
		RawLog:    res.DeliverTx.Log,
		Logs:      parsedLogs,
		Info:      res.DeliverTx.Info,
		GasWanted: res.DeliverTx.GasWanted,
		GasUsed:   res.DeliverTx.GasUsed,
		Events:    res.DeliverTx.Events,
	}
}

func newResponseFormatBroadcastTxCommit(res *coretypes.ResultBroadcastTxCommit) *sdk.TxResponse {
	if res == nil {
		return nil
	}

	if !res.CheckTx.IsOK() {
		return newTxResponseCheckTx(res)
	}

	return newTxResponseDeliverTx(res)
}

// QueryEventForTxCmd returns a CLI command that subscribes to a WebSocket connection and waits for a transaction event with the given hash.
func QueryEventForTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event-query-tx-for [hash]",
		Short: "Query for a transaction by hash",
		Long:  `Subscribes to a CometBFT WebSocket connection and waits for a transaction event with the given hash.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			hash := args[0]

			// XXX: We use a hardcoded 15 second timeout for compatibility with v0.47,
			// but this will be configurable sometime in v0.50.
			// You can use Agoric's -bblock if you don't want any timeout.
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
			defer cancel()

			// We block here until the event is received, but only if the above
			// context timeout doesn't happen first.
			waitTx := <-client.WaitTx(ctx, clientCtx.NodeURI, hash)
			if evt := waitTx.BlockInclusion; evt != nil {
				// There was block inclusion, so print the event.
				res := &coretypes.ResultBroadcastTxCommit{
					DeliverTx: evt.Result,
					Hash:      tmtypes.Tx(evt.Tx).Hash(),
					Height:    evt.Height,
				}

				err = clientCtx.PrintProto(newResponseFormatBroadcastTxCommit(res))
			}

			// Check for waiting errors, if any.
			if waitTx.Err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					return errors.ErrLogic.Wrapf("timed out waiting for event, the transaction could have already been included or wasn't yet included")
				}
				return waitTx.Err
			}

			// Propagate the printing error, if any.
			return err
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
