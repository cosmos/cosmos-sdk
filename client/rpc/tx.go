package rpc

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
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
<<<<<<< HEAD
		Use:   "event-query-tx-for [hash]",
		Short: "Query for a transaction by hash",
		Long:  `Subscribes to a CometBFT WebSocket connection and waits for a transaction event with the given hash.`,
		Args:  cobra.ExactArgs(1),
=======
		Use:     "wait-tx [hash]",
		Aliases: []string{"event-query-tx-for"},
		Short:   "Wait for a transaction to be included in a block",
		Long:    `Subscribes to a CometBFT WebSocket connection and waits for a transaction event with the given hash.`,
		Example: fmt.Sprintf(`By providing the transaction hash:
$ %[1]s q wait-tx [hash]

Or, by piping a "tx" command:
$ %[1]s tx [flags] | %[1]s q wait-tx
`, version.AppName),
		Args: cobra.MaximumNArgs(1),
>>>>>>> 55b00938b (fix: Properly parse json in the wait-tx command. (#20631))
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			c, err := rpchttp.New(clientCtx.NodeURI, "/websocket")
			if err != nil {
				return err
			}
			if err := c.Start(); err != nil {
				return err
			}
			defer c.Stop() //nolint:errcheck // ignore stop error

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
			defer cancel()

			hash := args[0]
			query := fmt.Sprintf("%s='%s' AND %s='%s'", tmtypes.EventTypeKey, tmtypes.EventTx, tmtypes.TxHashKey, hash)
			const subscriber = "subscriber"
			eventCh, err := c.Subscribe(ctx, subscriber, query)
			if err != nil {
				return fmt.Errorf("failed to subscribe to tx: %w", err)
			}
			defer c.UnsubscribeAll(context.Background(), subscriber) //nolint:errcheck // ignore unsubscribe error

			select {
			case evt := <-eventCh:
				if txe, ok := evt.Data.(tmtypes.EventDataTx); ok {
					res := &coretypes.ResultBroadcastTxCommit{
						DeliverTx: txe.Result,
						Hash:      tmtypes.Tx(txe.Tx).Hash(),
						Height:    txe.Height,
					}
					return clientCtx.PrintProto(newResponseFormatBroadcastTxCommit(res))
				}
			case <-ctx.Done():
				return errors.ErrLogic.Wrapf("timed out waiting for event, the transaction could have already been included or wasn't yet included")
			}
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
<<<<<<< HEAD
=======

func parseHashFromInput(in []byte) ([]byte, error) {
	// The content of in is expected to be the result of a tx command which should be using GenerateOrBroadcastTxCLI.
	// That outputs a sdk.TxResponse as either the json or yaml. As json, we can't unmarshal it back into that struct,
	// though, because the height field ends up quoted which confuses json.Unmarshal (because it's for an int64 field).

	// Try to find the txhash from json ouptut.
	resultTx := make(map[string]json.RawMessage)
	if err := json.Unmarshal(in, &resultTx); err == nil && len(resultTx["txhash"]) > 0 {
		// input was JSON, return the hash
		hash := strings.Trim(strings.TrimSpace(string(resultTx["txhash"])), `"`)
		if len(hash) > 0 {
			return hex.DecodeString(hash)
		}
	}

	// Try to find the txhash from yaml output.
	lines := strings.Split(string(in), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "txhash:") {
			hash := strings.TrimSpace(line[len("txhash:"):])
			return hex.DecodeString(hash)
		}
	}
	return nil, fmt.Errorf("txhash not found")
}
>>>>>>> 55b00938b (fix: Properly parse json in the wait-tx command. (#20631))
