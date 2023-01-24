package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/sanction"
)

// exampleQueryCmdBase is the base command that gets a user to one of the query commands in here.
var exampleQueryCmdBase = fmt.Sprintf("%s query %s", version.AppName, sanction.ModuleName)

var exampleQueryAddr1 = sdk.AccAddress("exampleQueryAddr1___")

// QueryCmd returns the command with sub-commands for specific sanction module queries.
func QueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        sanction.ModuleName,
		Short:                      "Querying commands for the sanction module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	queryCmd.AddCommand(
		QueryIsSanctionedCmd(),
		QuerySanctionedAddressesCmd(),
		QueryTemporaryEntriesCmd(),
		QueryParamsCmd(),
	)

	return queryCmd
}

// QueryIsSanctionedCmd returns the command for executing an IsSanctioned query.
func QueryIsSanctionedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "is-sanctioned <address>",
		Aliases: []string{"is", "check", "is-sanction", "c"},
		Short:   "Check if an address is sanctioned",
		Long: fmt.Sprintf(`Check if an address is sanctioned.

Examples:
  $ %[1]s is-sanctioned %[2]s
  $ %[1]s is %[2]s
  $ %[1]s check %[2]s
  $ %[1]s c %[2]s
`,
			exampleQueryCmdBase, exampleQueryAddr1),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			if _, err = sdk.AccAddressFromBech32(args[0]); err != nil {
				return sdkerrors.ErrInvalidAddress.Wrap(err.Error())
			}

			req := sanction.QueryIsSanctionedRequest{
				Address: args[0],
			}

			var res *sanction.QueryIsSanctionedResponse
			queryClient := sanction.NewQueryClient(clientCtx)
			res, err = queryClient.IsSanctioned(cmd.Context(), &req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QuerySanctionedAddressesCmd returns a command for executing a SanctionedAddresses query.
func QuerySanctionedAddressesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sanctioned-addresses",
		Aliases: []string{"addresses", "all"},
		Short:   "List all the sanctioned addresses",
		Long: fmt.Sprintf(`List all the sanctioned addresses.

Examples:
  $ %[1]s sanctioned-addresses
  $ %[1]s addresses
  $ %[1]s all
`,
			exampleQueryCmdBase),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			req := sanction.QuerySanctionedAddressesRequest{}
			req.Pagination, err = client.ReadPageRequestWithPageKeyDecoded(cmd.Flags())
			if err != nil {
				return err
			}

			var res *sanction.QuerySanctionedAddressesResponse
			queryClient := sanction.NewQueryClient(clientCtx)
			res, err = queryClient.SanctionedAddresses(cmd.Context(), &req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "sanctioned-addresses")

	return cmd
}

// QueryTemporaryEntriesCmd returns a command for executing a TemporaryEntries query.
func QueryTemporaryEntriesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "temporary-entries [<address>]",
		Aliases: []string{"temp-entries", "temp"},
		Short:   "List temporarily sanctioned/unsanctioned addresses",
		Long: fmt.Sprintf(`List all temporarily sanctioned/unsanctioned addresses.
If an address is provided, only temporary entries for that address are returned.
Otherwise, all temporary entries are returned.

Examples:
  $ %[1]s temporary-entries
  $ %[1]s temporary-entries %[2]s
  $ %[1]s temp-entries
  $ %[1]s temp-entries %[2]s
  $ %[1]s temp
  $ %[1]s temp %[2]s
`,
			exampleQueryCmdBase, exampleQueryAddr1),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			if len(args) > 0 {
				if _, err := sdk.AccAddressFromBech32(args[0]); err != nil {
					return err
				}
			}

			req := sanction.QueryTemporaryEntriesRequest{}
			if len(args) > 0 {
				req.Address = args[0]
			}

			req.Pagination, err = client.ReadPageRequestWithPageKeyDecoded(cmd.Flags())
			if err != nil {
				return err
			}

			var res *sanction.QueryTemporaryEntriesResponse
			queryClient := sanction.NewQueryClient(clientCtx)
			res, err = queryClient.TemporaryEntries(cmd.Context(), &req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "temporary-entries")

	return cmd
}

// QueryParamsCmd returns a command for executing a Params query.
func QueryParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Get the sanction module params",
		Long: fmt.Sprintf(`Get the sanction module params.

Example:
  $ %[1]s params
`,
			exampleQueryCmdBase),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			req := sanction.QueryParamsRequest{}

			var res *sanction.QueryParamsResponse
			queryClient := sanction.NewQueryClient(clientCtx)
			res, err = queryClient.Params(cmd.Context(), &req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
