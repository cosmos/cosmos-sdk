package cli

import (
	"context"
	"fmt"
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
		flags.GetCommands(
			GetCmdQuerySigningInfo(),
			GetCmdQueryParams(),
		)...,
	)

	return slashingQueryCmd

}

// GetCmdQuerySigningInfo implements the command to query signing info.
func GetCmdQuerySigningInfo() *cobra.Command {
	return &cobra.Command{
		Use:   "signing-info [validator-conspub]",
		Short: "Query a validator's signing information",
		Long: strings.TrimSpace(`Use a validators' consensus public key to find the signing-info for that validator:

$ <appcli> query slashing signing-info cosmosvalconspub1zcjduepqfhvwcmt7p06fvdgexxhmz0l8c7sgswl7ulv7aulk364x4g5xsw7sr0k2g5
`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			pk, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, args[0])
			if err != nil {
				return err
			}

			consAddr := sdk.ConsAddress(pk.Address())
			params := &types.QuerySigningInfoRequest{ConsAddress: consAddr}
			res, err := queryClient.SigningInfo(context.Background(), params)

			if err != nil {
				return err
			}

			if res.String() == "" {
				return fmt.Errorf("validator %s not found in slashing store", consAddr)
			}

			return clientCtx.PrintOutput(res.ValSigningInfo)
		},
	}
}

// GetCmdQueryParams implements a command to fetch slashing parameters.
func GetCmdQueryParams() *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "Query the current slashing parameters",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(`Query genesis parameters for the slashing module:

$ <appcli> query slashing params
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryParamsRequest{}
			res, err := queryClient.Params(context.Background(), params)

			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res.Params)
		},
	}
}
