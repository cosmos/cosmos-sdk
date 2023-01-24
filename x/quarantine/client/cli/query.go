package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
)

// exampleQueryCmdBase is the base command that gets a user to one of the query commands in here.
var exampleQueryCmdBase = fmt.Sprintf("%s query %s", version.AppName, quarantine.ModuleName)

// QueryCmd returns the command with sub-commands for specific quarantine module queries.
func QueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        quarantine.ModuleName,
		Short:                      "Querying commands for the quarantine module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	queryCmd.AddCommand(
		QueryQuarantinedFundsCmd(),
		QueryIsQuarantinedCmd(),
		QueryAutoResponsesCmd(),
	)

	return queryCmd
}

// QueryQuarantinedFundsCmd returns the command for executing a QuarantinedFunds query.
func QueryQuarantinedFundsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "funds [<to_address> [<from_address>]]",
		Short: "Query for quarantined funds",
		Long: fmt.Sprintf(`Query for quarantined funds.

If no arguments are provided, all quarantined funds will be returned.
If only a to_address is provided, only undeclined funds quarantined for that address are returned.
If both a to_address and from_address are provided, quarantined funds will be returned regardless of whether they've been declined.

Examples:
  $ %[1]s funds
  $ %[1]s funds %[2]s
  $ %[1]s funds %[2]s %[3]s
`,
			exampleQueryCmdBase, exampleAddr1, exampleAddr2),
		Args: cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			req := quarantine.QueryQuarantinedFundsRequest{}

			if len(args) > 0 {
				req.ToAddress, err = validateAddress(args[0], "to_address")
				if err != nil {
					return err
				}
			}

			if len(args) > 1 {
				req.FromAddress, err = validateAddress(args[1], "from_address")
				if err != nil {
					return err
				}
			}

			req.Pagination, err = client.ReadPageRequestWithPageKeyDecoded(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := quarantine.NewQueryClient(clientCtx)

			var res *quarantine.QueryQuarantinedFundsResponse
			res, err = queryClient.QuarantinedFunds(cmd.Context(), &req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "quarantined funds")

	return cmd
}

// QueryIsQuarantinedCmd returns the command for executing an IsQuarantined query.
func QueryIsQuarantinedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "is-quarantined <to_address>",
		Aliases: []string{"is", "check", "c"},
		Short:   "Query whether an account is opted into quarantined",
		Long: fmt.Sprintf(`Query whether an account is opted into quarantined.

Examples:
  $ %[1]s is-quarantined %[2]s
  $ %[1]s is %[2]s
  $ %[1]s check %[2]s
  $ %[1]s c %[2]s
`,
			exampleQueryCmdBase, exampleAddr1),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			req := quarantine.QueryIsQuarantinedRequest{}

			req.ToAddress, err = validateAddress(args[0], "to_address")
			if err != nil {
				return err
			}

			queryClient := quarantine.NewQueryClient(clientCtx)

			var res *quarantine.QueryIsQuarantinedResponse
			res, err = queryClient.IsQuarantined(cmd.Context(), &req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryAutoResponsesCmd returns the command for executing a AutoResponses query.
func QueryAutoResponsesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "auto-responses <to_address> [<from_address>]",
		Aliases: []string{"auto", "ar"},
		Short:   "Query auto-responses",
		Long: fmt.Sprintf(`Query auto-responses.

If only a to_address is provided, all auto-responses set up for that address are returned. This will only contain accept or decline entries.
If both a to_address and from_address are provided, exactly one result will be returned. This can be accept, decline or unspecified.

Examples:
  $ %[1]s auto-responses %[2]s
  $ %[1]s auto-responses %[2]s %[3]s
`,
			exampleQueryCmdBase, exampleAddr1, exampleAddr2),
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			req := quarantine.QueryAutoResponsesRequest{}

			req.ToAddress, err = validateAddress(args[0], "to_address")
			if err != nil {
				return err
			}

			if len(args) > 1 {
				req.FromAddress, err = validateAddress(args[1], "from_address")
				if err != nil {
					return err
				}
			}

			req.Pagination, err = client.ReadPageRequestWithPageKeyDecoded(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := quarantine.NewQueryClient(clientCtx)

			var res *quarantine.QueryAutoResponsesResponse
			res, err = queryClient.AutoResponses(cmd.Context(), &req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "auto-responses")

	return cmd
}
