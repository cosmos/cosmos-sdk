package cli

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	// Group slashing queries under a subcommand
	slashingQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the slashing module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	slashingQueryCmd.AddCommand(
		GetCmdQuerySigningInfo(),
		GetCmdQueryParams(),
		GetCmdQuerySigningInfos(),
	)

	return slashingQueryCmd

}

// GetCmdQuerySigningInfo implements the command to query signing info.
func GetCmdQuerySigningInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signing-info [validator-conspub]",
		Short: "Query a validator's signing information",
		Long: strings.TrimSpace(`Use a validators' consensus public key to find the signing-info for that validator:

$ <appd> query slashing signing-info cosmosvalconspub1zcjduepqfhvwcmt7p06fvdgexxhmz0l8c7sgswl7ulv7aulk364x4g5xsw7sr0k2g5
`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			pk, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, args[0])
			if err != nil {
				return err
			}

			consAddr := sdk.ConsAddress(pk.Address())
			params := &types.QuerySigningInfoRequest{ConsAddress: consAddr.String()}
			res, err := queryClient.SigningInfo(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.ValSigningInfo)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQuerySigningInfos implements the command to query signing infos.
func GetCmdQuerySigningInfos() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signing-infos",
		Short: "Query signing information of all validators",
		Long: strings.TrimSpace(`signing infos of validators:

$ <appd> query slashing signing-infos
`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			params := &types.QuerySigningInfosRequest{Pagination: pageReq}
			res, err := queryClient.SigningInfos(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "signing infos")

	return cmd
}

// GetCmdQueryParams implements a command to fetch slashing parameters.
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current slashing parameters",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(`Query genesis parameters for the slashing module:

$ <appd> query slashing params
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryParamsRequest{}
			res, err := queryClient.Params(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
