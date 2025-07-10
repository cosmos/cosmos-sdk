package rpc

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
)

const TimeoutFlag = "timeout"

// QueryEventForTxCmd is an alias for WaitTxCmd, kept for backwards compatibility.
func QueryEventForTxCmd() *cobra.Command {
	return WaitTxCmd()
}

// WaitTxCmd returns a CLI command that waits for a transaction with the given hash to be included in a block.
func WaitTxCmd() *cobra.Command {
	cmd := &cobra.Command{
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
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			timeout, err := cmd.Flags().GetDuration(TimeoutFlag)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			txSubmitted := func() (*sdk.TxResponse, []byte, error) {
				if len(args) > 0 {
					// read hash from args
					hashByt, err := hex.DecodeString(args[0])
					if err != nil {
						return nil, nil, err
					}

					return nil, hashByt, nil
				}

				// read hash from stdin
				in, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return nil, nil, err
				}
				hashByt, err := parseHashFromInput(in)
				if err != nil {
					return nil, nil, err
				}

				return nil, hashByt, nil
			}

			res, err := clientCtx.WaitTx(ctx, txSubmitted)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().Duration(TimeoutFlag, 15*time.Second, "The maximum time to wait for the transaction to be included in a block")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

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
