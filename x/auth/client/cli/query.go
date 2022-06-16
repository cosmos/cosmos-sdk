package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/version"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
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
		GetAccountAddressByIDCmd(),
		GetAccountsCmd(),
		QueryParamsCmd(),
		QueryModuleAccountByNameCmd(),
	)

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
		Use:   "tx --type={hash|acc_seq|signature} <hash|acc_seq|signature>",
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
					return errors.New("argument should be a tx hash")
				}

				// if hash is given, then query the tx by hash
				output, err := authtx.QueryTx(clientCtx, args[0])
				if err != nil {
					return err
				}

				if output.Empty() {
					return fmt.Errorf("no transaction found with hash %s", args[0])
				}

// GetAccountAddressByIDCmd returns a query account that will display the account address of a given account id.
func GetAccountAddressByIDCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "address-by-id [id]",
		Short:   "Query for account address by account id",
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf("%s q auth address-by-id 1", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.AccountAddressByID(cmd.Context(), &types.QueryAccountAddressByIDRequest{Id: id})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetAccountsCmd returns a query command that will display a list of accounts
func GetAccountsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accounts",
		Short: "Query all the accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

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
					return errors.New("found no txs matching given signatures")
				}
				if len(txs.Txs) > 1 {
					// This case means there's a bug somewhere else in the code as this
					// should not happen.
					return sdkerrors.ErrLogic.Wrapf("found %d txs matching given signatures", len(txs.Txs))
				}

// QueryModuleAccountByNameCmd returns a command to
func QueryModuleAccountByNameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "module-account [module-name]",
		Short:   "Query module account info by module name",
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf("%s q auth module-account auth", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			moduleName := args[0]
			if len(moduleName) == 0 {
				return fmt.Errorf("module name should not be empty")
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.ModuleAccountByName(context.Background(), &types.QueryModuleAccountByNameRequest{Name: moduleName})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
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

			case TypeAccSeq:
				if args[0] == "" {
					return errors.New("`acc_seq` type takes an argument '<addr>/<seq>'")
				}

				query := fmt.Sprintf("%s.%s='%s'", sdk.EventTypeTx, sdk.AttributeKeyAccountSequence, args[0])

				txs, err := authtx.QueryTxsByEvents(clientCtx, querytypes.DefaultPage, querytypes.DefaultLimit, query, "")
				if err != nil {
					return err
				}

				if len(txs.Txs) == 0 {
					return errors.New("found no txs matching given address and sequence combination")
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
					output, err := authtx.QueryTx(clientCtx, args[0])
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

					txs, err := authtx.QueryTxsByEvents(clientCtx, tmEvents, rest.DefaultPage, query.DefaultLimit, "")
					if err != nil {
						return err
					}
					if len(txs.Txs) == 0 {
						return fmt.Errorf("found no txs matching given signatures")
					}
					if len(txs.Txs) > 1 {
						// This case means there's a bug somewhere else in the code. Should not happen.
						return errors.ErrLogic.Wrapf("found %d txs matching given signatures", len(txs.Txs))
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
					txs, err := authtx.QueryTxsByEvents(clientCtx, tmEvents, rest.DefaultPage, query.DefaultLimit, "")
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
