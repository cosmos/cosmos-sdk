package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

// GetQueryCmd returns the cli query commands for the minting module.
func GetQueryCmd() *cobra.Command {
	mintingQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the minting module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	mintingQueryCmd.AddCommand(
		flags.GetCommands(
			GetCmdQueryParams(),
			GetCmdQueryInflation(),
			GetCmdQueryAnnualProvisions(),
		)...,
	)

	return mintingQueryCmd
}

// GetCmdQueryParams implements a command to return the current minting
// parameters.
func GetCmdQueryParams() *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "Query the current minting parameters",
		Args:  cobra.NoArgs,
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

// GetCmdQueryInflation implements a command to return the current minting
// inflation value.
func GetCmdQueryInflation() *cobra.Command {
	return &cobra.Command{
		Use:   "inflation",
		Short: "Query the current minting inflation value",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryInflationRequest{}
			res, err := queryClient.Inflation(context.Background(), params)

			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res.Inflation)
		},
	}
}

// GetCmdQueryAnnualProvisions implements a command to return the current minting
// annual provisions value.
func GetCmdQueryAnnualProvisions() *cobra.Command {
	return &cobra.Command{
		Use:   "annual-provisions",
		Short: "Query the current minting annual provisions value",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAnnualProvisionsRequest{}
			res, err := queryClient.AnnualProvisions(context.Background(), params)

			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res.AnnualProvisions)
		},
	}
}
