package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/version"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

const (
	flagEvents = "events"
	flagType   = "type"

	typeHash   = "hash"
	typeAccSeq = "acc_seq"
	typeSig    = "signature"

	eventFormat = "{eventType}.{eventAttribute}={value}"
)

// GetQueryCmd returns the transaction commands for this module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the auth module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetAccountCmd(),
		QueryParamsCmd(),
	)

	return cmd
}

// QueryParamsCmd returns the command handler for evidence parameter querying.
func QueryParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current auth parameters",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(`Query the current auth parameters:

$ <appd> query auth params
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetAccountCmd returns a query account that will display the state of the
// account at a given address.
func GetAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account [address]",
		Short: "Query for account by address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			key, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Account(context.Background(), &types.QueryAccountRequest{Address: key.String()})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res.Account)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryTxsByEventsCmd returns a command to search through transactions by events.
func QueryTxsByEventsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txs",
		Short: "Query for paginated transactions that match a set of events",
		Long: strings.TrimSpace(
			fmt.Sprintf(`
Search for transactions that match the exact given events where results are paginated.
Each event takes the form of '%s'. Please refer
to each module's documentation for the full set of events to query for. Each module
documents its respective events under 'xx_events.md'.

Example:
$ %s query txs --%s 'message.sender=cosmos1...&message.action=withdraw_delegator_reward' --page 1 --limit 30
`, eventFormat, version.AppName, flagEvents),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			eventsRaw, _ := cmd.Flags().GetString(flagEvents)
			eventsStr := strings.Trim(eventsRaw, "'")

			var events []string
			if strings.Contains(eventsStr, "&") {
				events = strings.Split(eventsStr, "&")
			} else {
				events = append(events, eventsStr)
			}

			var tmEvents []string

			for _, event := range events {
				if !strings.Contains(event, "=") {
					return fmt.Errorf("invalid event; event %s should be of the format: %s", event, eventFormat)
				} else if strings.Count(event, "=") > 1 {
					return fmt.Errorf("invalid event; event %s should be of the format: %s", event, eventFormat)
				}

				tokens := strings.Split(event, "=")
				if tokens[0] == tmtypes.TxHeightKey {
					event = fmt.Sprintf("%s=%s", tokens[0], tokens[1])
				} else {
					event = fmt.Sprintf("%s='%s'", tokens[0], tokens[1])
				}

				tmEvents = append(tmEvents, event)
			}

			page, _ := cmd.Flags().GetInt(flags.FlagPage)
			limit, _ := cmd.Flags().GetInt(flags.FlagLimit)

			txs, err := authclient.QueryTxsByEvents(clientCtx, tmEvents, page, limit, "")
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(txs)
		},
	}

	cmd.Flags().StringP(flags.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")
	cmd.Flags().Int(flags.FlagPage, rest.DefaultPage, "Query a specific page of paginated results")
	cmd.Flags().Int(flags.FlagLimit, rest.DefaultLimit, "Query number of transactions results per page returned")
	cmd.Flags().String(flagEvents, "", fmt.Sprintf("list of transaction events in the form of %s", eventFormat))
	cmd.MarkFlagRequired(flagEvents)

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
			version.AppName, flagType, typeAccSeq,
			version.AppName, flagType, typeSig)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			typ, _ := cmd.Flags().GetString(flagType)

			switch typ {
			case typeHash:
				{
					if args[0] == "" {
						return fmt.Errorf("argument should be a tx hash")
					}

					// If hash is given, then query the tx by hash.
					output, err := authclient.QueryTx(clientCtx, args[0])
					if err != nil {
						return err
					}

					if output.Empty() {
						return fmt.Errorf("no transaction found with hash %s", args[0])
					}

					return clientCtx.PrintProto(output)
				}
			case typeSig:
				{
					sigParts, err := parseSigArgs(args)
					if err != nil {
						return err
					}
					tmEvents := make([]string, len(sigParts))
					for i, sig := range sigParts {
						tmEvents[i] = fmt.Sprintf("%s.%s='%s'", sdk.EventTypeTx, sdk.AttributeKeySignature, sig)
					}

					txs, err := authclient.QueryTxsByEvents(clientCtx, tmEvents, rest.DefaultPage, query.DefaultLimit, "")
					if err != nil {
						return err
					}
					if len(txs.Txs) == 0 {
						return fmt.Errorf("found no txs matching given signatures")
					}
					if len(txs.Txs) > 1 {
						// This case means there's a bug somewhere else in the code. Should not happen.
						return errors.Wrapf(errors.ErrLogic, "found %d txs matching given signatures", len(txs.Txs))
					}

					return clientCtx.PrintProto(txs.Txs[0])
				}
			case typeAccSeq:
				{
					if args[0] == "" {
						return fmt.Errorf("`acc_seq` type takes an argument '<addr>/<seq>'")
					}

					tmEvents := []string{
						fmt.Sprintf("%s.%s='%s'", sdk.EventTypeTx, sdk.AttributeKeyAccountSequence, args[0]),
					}
					txs, err := authclient.QueryTxsByEvents(clientCtx, tmEvents, rest.DefaultPage, query.DefaultLimit, "")
					if err != nil {
						return err
					}
					if len(txs.Txs) == 0 {
						return fmt.Errorf("found no txs matching given address and sequence combination")
					}
					if len(txs.Txs) > 1 {
						// This case means there's a bug somewhere else in the code. Should not happen.
						return fmt.Errorf("found %d txs matching given address and sequence combination", len(txs.Txs))
					}

					return clientCtx.PrintProto(txs.Txs[0])
				}
			default:
				return fmt.Errorf("unknown --%s value %s", flagType, typ)
			}
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().String(flagType, typeHash, fmt.Sprintf("The type to be used when querying tx, can be one of \"%s\", \"%s\", \"%s\"", typeHash, typeAccSeq, typeSig))

	return cmd
}

// parseSigArgs parses comma-separated signatures from the CLI arguments.
func parseSigArgs(args []string) ([]string, error) {
	if len(args) != 1 || args[0] == "" {
		return nil, fmt.Errorf("argument should be comma-separated signatures")
	}

	return strings.Split(args[0], ","), nil
}
