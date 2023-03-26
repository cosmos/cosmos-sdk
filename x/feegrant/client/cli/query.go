package cli

import (
	"fmt"
	"strings"

	"cosmossdk.io/core/address"
	"github.com/spf13/cobra"

	"cosmossdk.io/x/feegrant"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(ac address.Codec) *cobra.Command {
	feegrantQueryCmd := &cobra.Command{
		Use:                        feegrant.ModuleName,
		Short:                      "Querying commands for the feegrant module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	feegrantQueryCmd.AddCommand(
		GetCmdQueryFeeGrant(ac),
		GetCmdQueryFeeGrantsByGrantee(ac),
		GetCmdQueryFeeGrantsByGranter(ac),
	)

	return feegrantQueryCmd
}

// GetCmdQueryFeeGrant returns cmd to query for a grant between granter and grantee.
func GetCmdQueryFeeGrant(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grant [granter] [grantee]",
		Args:  cobra.ExactArgs(2),
		Short: "Query details of a single grant",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details for a grant. 
You can find the fee-grant of a granter and grantee.

Example:
$ %s query feegrant grant [granter] [grantee]
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := feegrant.NewQueryClient(clientCtx)

			if _, err := ac.StringToBytes(args[0]); err != nil {
				return err
			}

			if _, err := ac.StringToBytes(args[1]); err != nil {
				return err
			}

			res, err := queryClient.Allowance(
				cmd.Context(),
				&feegrant.QueryAllowanceRequest{
					Granter: args[0],
					Grantee: args[1],
				},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res.Allowance)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryFeeGrantsByGrantee returns cmd to query for all grants for a grantee.
func GetCmdQueryFeeGrantsByGrantee(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grants-by-grantee [grantee]",
		Args:  cobra.ExactArgs(1),
		Short: "Query all grants of a grantee",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries all the grants for a grantee address.

Example:
$ %s query feegrant grants-by-grantee [grantee]
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := feegrant.NewQueryClient(clientCtx)

			_, err := ac.StringToBytes(args[0])
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.Allowances(
				cmd.Context(),
				&feegrant.QueryAllowancesRequest{
					Grantee:    args[0],
					Pagination: pageReq,
				},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "grants")

	return cmd
}

// GetCmdQueryFeeGrantsByGranter returns cmd to query for all grants by a granter.
func GetCmdQueryFeeGrantsByGranter(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grants-by-granter [granter]",
		Args:  cobra.ExactArgs(1),
		Short: "Query all grants by a granter",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries all the grants issued for a granter address.

Example:
$ %s query feegrant grants-by-granter [granter]
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := feegrant.NewQueryClient(clientCtx)

			_, err := ac.StringToBytes(args[0])
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.AllowancesByGranter(
				cmd.Context(),
				&feegrant.QueryAllowancesByGranterRequest{
					Granter:    args[0],
					Pagination: pageReq,
				},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "grants")

	return cmd
}
