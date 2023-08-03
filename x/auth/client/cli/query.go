package cli

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/version"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
)

const (
	FlagQuery   = "query"
	FlagType    = "type"
	FlagOrderBy = "order_by"

	TypeHash   = "hash"
	TypeAccSeq = "acc_seq"
	TypeSig    = "signature"
	TypeHeight = "height"

	EventFormat = "{eventType}.{eventAttribute}={value}"
)

// QueryTxsByEventsCmd returns a command to search through transactions by events.
func QueryTxsByEventsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txs",
		Short: "Query for paginated transactions that match a set of events",
		Long: `Search for transactions that match the exact given events where results are paginated.
The events query is directly passed to Tendermint's RPC TxSearch method and must
conform to Tendermint's query syntax.

Please refer to each module's documentation for the full set of events to query
for. Each module documents its respective events under 'xx_events.md'.
`,
		Example: fmt.Sprintf(
			"$ %s query txs --query \"message.sender='cosmos1...' AND message.action='withdraw_delegator_reward' AND tx.height > 7\" --page 1 --limit 30",
			version.AppName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			query, _ := cmd.Flags().GetString(FlagQuery)
			page, _ := cmd.Flags().GetInt(flags.FlagPage)
			limit, _ := cmd.Flags().GetInt(flags.FlagLimit)
			orderBy, _ := cmd.Flags().GetString(FlagOrderBy)

			txs, err := authtx.QueryTxsByEvents(clientCtx, page, limit, query, orderBy)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(txs)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().Int(flags.FlagPage, querytypes.DefaultPage, "Query a specific page of paginated results")
	cmd.Flags().Int(flags.FlagLimit, querytypes.DefaultLimit, "Query number of transactions results per page returned")
	cmd.Flags().String(FlagQuery, "", "The transactions events query per Tendermint's query semantics")
	cmd.Flags().String(FlagOrderBy, "", "The ordering semantics (asc|dsc)")
	_ = cmd.MarkFlagRequired(FlagQuery)

	return cmd
}

// QueryTxCmd implements the default command for a tx query.
func QueryTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx --type=[hash|acc_seq|signature] [hash|acc_seq|signature]",
		Short: "Query for a transaction by hash, \"<addr>/<seq>\" combination or comma-separated signatures in a committed block",
		Long: strings.TrimSpace(fmt.Sprintf(`
Example:
$ %s query tx <hash>
$ %s query tx --%s=%s <addr>/<sequence>
$ %s query tx --%s=%s <sig1_base64>,<sig2_base64...>
`,
			version.AppName,
			version.AppName, FlagType, TypeAccSeq,
			version.AppName, FlagType, TypeSig)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			typ, _ := cmd.Flags().GetString(FlagType)

			switch typ {
			case TypeHash:
				if args[0] == "" {
					return fmt.Errorf("argument should be a tx hash")
				}

				// if hash is given, then query the tx by hash
				output, err := authtx.QueryTx(clientCtx, args[0])
				if err != nil {
					return err
				}

				if output.Empty() {
					return fmt.Errorf("no transaction found with hash %s", args[0])
				}

				return clientCtx.PrintProto(output)

			case TypeSig:
				sigParts, err := ParseSigArgs(args)
				if err != nil {
					return err
				}

				events := make([]string, len(sigParts))
				for i, sig := range sigParts {
					events[i] = fmt.Sprintf("%s.%s='%s'", sdk.EventTypeTx, sdk.AttributeKeySignature, sig)
				}

				query := strings.Join(events, " AND ")

				txs, err := authtx.QueryTxsByEvents(clientCtx, querytypes.DefaultPage, querytypes.DefaultLimit, query, "")
				if err != nil {
					return err
				}

				if len(txs.Txs) == 0 {
					return fmt.Errorf("found no txs matching given signatures")
				}
				if len(txs.Txs) > 1 {
					// This case means there's a bug somewhere else in the code as this
					// should not happen.
					return errors.ErrLogic.Wrapf("found %d txs matching given signatures", len(txs.Txs))
				}

				return clientCtx.PrintProto(txs.Txs[0])

			case TypeAccSeq:
				if args[0] == "" {
					return fmt.Errorf("`acc_seq` type takes an argument '<addr>/<seq>'")
				}

				query := fmt.Sprintf("%s.%s='%s'", sdk.EventTypeTx, sdk.AttributeKeyAccountSequence, args[0])

				txs, err := authtx.QueryTxsByEvents(clientCtx, querytypes.DefaultPage, querytypes.DefaultLimit, query, "")
				if err != nil {
					return err
				}

				if len(txs.Txs) == 0 {
					return fmt.Errorf("found no txs matching given address and sequence combination")
				}
				if len(txs.Txs) > 1 {
					// This case means there's a bug somewhere else in the code as this
					// should not happen.
					return fmt.Errorf("found %d txs matching given address and sequence combination", len(txs.Txs))
				}

				return clientCtx.PrintProto(txs.Txs[0])

			default:
				return fmt.Errorf("unknown --%s value %s", FlagType, typ)
			}
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().String(FlagType, TypeHash, fmt.Sprintf("The type to be used when querying tx, can be one of \"%s\", \"%s\", \"%s\"", TypeHash, TypeAccSeq, TypeSig))

	return cmd
}

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

	parsedLogs, _ := sdk.ParseABCILogs(res.TxResult.Log)

	return &sdk.TxResponse{
		Height:    res.Height,
		TxHash:    txHash,
		Codespace: res.TxResult.Codespace,
		Code:      res.TxResult.Code,
		Data:      strings.ToUpper(hex.EncodeToString(res.TxResult.Data)),
		RawLog:    res.TxResult.Log,
		Logs:      parsedLogs,
		Info:      res.TxResult.Info,
		GasWanted: res.TxResult.GasWanted,
		GasUsed:   res.TxResult.GasUsed,
		Events:    res.TxResult.Events,
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
						TxResult: txe.Result,
						Hash:     tmtypes.Tx(txe.Tx).Hash(),
						Height:   txe.Height,
					}
					return clientCtx.PrintProto(newResponseFormatBroadcastTxCommit(res))
				}
			case <-ctx.Done():
				return errors.ErrLogic.Wrapf("timed out waiting for event")
			}
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// ParseSigArgs parses comma-separated signatures from the CLI arguments.
func ParseSigArgs(args []string) ([]string, error) {
	if len(args) != 1 || args[0] == "" {
		return nil, fmt.Errorf("argument should be comma-separated signatures")
	}

	return strings.Split(args[0], ","), nil
}
