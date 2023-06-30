package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/version"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

const (
	FlagEvents  = "events" // TODO: Remove when #14758 is merged
	FlagQuery   = "query"
	FlagType    = "type"
	FlagOrderBy = "order_by"

	TypeHash   = "hash"
	TypeAccSeq = "acc_seq"
	TypeSig    = "signature"
	TypeHeight = "height"

	EventFormat = "{eventType}.{eventAttribute}={value}"
)

// GetQueryCmd returns the transaction commands for this module
func GetQueryCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the auth module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetAccountCmd(ac),
		GetAccountAddressByIDCmd(),
		GetAccountsCmd(),
		QueryParamsCmd(),
		QueryModuleAccountsCmd(),
		QueryModuleAccountByNameCmd(),
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
			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
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
func GetAccountCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account [address]",
		Short: "Query for account by address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			_, err = ac.StringToBytes(args[0])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Account(cmd.Context(), &types.QueryAccountRequest{Address: args[0]})
			if err != nil {
				node, err2 := clientCtx.GetNode()
				if err2 != nil {
					return err2
				}
				status, err2 := node.Status(context.Background())
				if err2 != nil {
					return err2
				}
				catchingUp := status.SyncInfo.CatchingUp
				if !catchingUp {
					return errorsmod.Wrapf(err, "your node may be syncing, please check node status using `/status`")
				}
				return err
			}

			return clientCtx.PrintProto(res.Account)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetAccountAddressByIDCmd returns a query account that will display the account address of a given account id.
func GetAccountAddressByIDCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "address-by-acc-num [acc-num]",
		Aliases: []string{"address-by-id"},
		Short:   "Query for an address by account number",
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf("%s q auth address-by-acc-num 1", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			accNum, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.AccountAddressByID(cmd.Context(), &types.QueryAccountAddressByIDRequest{
				AccountId: accNum,
			})
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

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Accounts(cmd.Context(), &types.QueryAccountsRequest{Pagination: pageReq})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "all-accounts")

	return cmd
}

// QueryAllModuleAccountsCmd returns a list of all the existing module accounts with their account information and permissions
func QueryModuleAccountsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module-accounts",
		Short: "Query all module accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.ModuleAccounts(context.Background(), &types.QueryModuleAccountsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
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

// ParseSigArgs parses comma-separated signatures from the CLI arguments.
func ParseSigArgs(args []string) ([]string, error) {
	if len(args) != 1 || args[0] == "" {
		return nil, fmt.Errorf("argument should be comma-separated signatures")
	}

	return strings.Split(args[0], ","), nil
}
