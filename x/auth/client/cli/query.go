package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/version"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

const (
	flagEvents = "events"

	eventFormat = "{eventType}.{eventAttribute}={value}"
)

// GetQueryCmd returns the transaction commands for this module
func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the auth module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetAccountCmd(cdc),
		QueryParamsCmd(cdc),
	)

	return cmd
}

// QueryParamsCmd returns the command handler for evidence parameter querying.
func QueryParamsCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current auth parameters",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(`Query the current auth parameters:

$ <appcli> query auth params
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.NewContext().WithCodec(cdc)

			route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryParams)
			res, _, err := clientCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			var params types.Params
			if err := cdc.UnmarshalJSON(res, &params); err != nil {
				return fmt.Errorf("failed to unmarshal params: %w", err)
			}

			return clientCtx.PrintOutput(params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetAccountCmd returns a query account that will display the state of the
// account at a given address.
func GetAccountCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account [address]",
		Short: "Query for account by address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.NewContext().WithCodec(cdc)
			accGetter := types.NewAccountRetriever(authclient.Codec)

			key, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			acc, err := accGetter.GetAccount(clientCtx, key)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(acc)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryTxsByEventsCmd returns a command to search through transactions by events.
func QueryTxsByEventsCmd(cdc *codec.Codec) *cobra.Command {
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

			clientCtx := client.NewContext().WithCodec(cdc)
			txs, err := authclient.QueryTxsByEvents(clientCtx, tmEvents, page, limit, "")
			if err != nil {
				return err
			}

			output, err := cdc.MarshalJSON(txs)
			if err != nil {
				return err
			}

			fmt.Println(string(output))
			return nil
		},
	}

	cmd.Flags().StringP(flags.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	cmd.Flags().Bool(flags.FlagTrustNode, false, "Trust connected full node (don't verify proofs for responses)")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")
	cmd.Flags().Int(flags.FlagPage, rest.DefaultPage, "Query a specific page of paginated results")
	cmd.Flags().Int(flags.FlagLimit, rest.DefaultLimit, "Query number of transactions results per page returned")
	cmd.Flags().String(flagEvents, "", fmt.Sprintf("list of transaction events in the form of %s", eventFormat))
	cmd.MarkFlagRequired(flagEvents)

	return cmd
}

// QueryTxCmd implements the default command for a tx query.
func QueryTxCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx [hash]",
		Short: "Query for a transaction by hash in a committed block",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.NewContext().WithCodec(cdc)

			output, err := authclient.QueryTx(clientCtx, args[0])
			if err != nil {
				return err
			}

			if output.Empty() {
				return fmt.Errorf("no transaction found with hash %s", args[0])
			}

			return clientCtx.PrintOutput(output)
		},
	}

	cmd.Flags().StringP(flags.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	cmd.Flags().Bool(flags.FlagTrustNode, false, "Trust connected full node (don't verify proofs for responses)")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")

	return cmd
}
