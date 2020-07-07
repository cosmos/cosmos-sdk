package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// GetCurrentPlanCmd returns the query upgrade plan command
func GetCurrentPlanCmd(clientCtx client.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "plan",
		Short: "get upgrade plan (if one exists)",
		Long:  "Gets the currently scheduled upgrade plan, if one exists",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			queryClient := types.NewQueryClient(clientCtx.Init())

			params := types.NewQueryCurrentPlanRequest()
			res, err := queryClient.CurrentPlan(context.Background(), params)
			if err != nil {
				return err
			}

			if len(res.Plan.Name) == 0 {
				return fmt.Errorf("no upgrade scheduled")
			}

			if err != nil {
				return err
			}
			return clientCtx.PrintOutput(res.Plan)
		},
	}
}

// GetAppliedPlanCmd returns the height at which a completed upgrade was applied
func GetAppliedPlanCmd(clientCtx client.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "applied [upgrade-name]",
		Short: "block header for height at which a completed upgrade was applied",
		Long: "If upgrade-name was previously executed on the chain, this returns the header for the block at which it was applied.\n" +
			"This helps a client determine which binary was valid over a given range of blocks, as well as more context to understand past migrations.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			queryClient := types.NewQueryClient(clientCtx.Init())

			name := args[0]
			params := types.NewQueryAppliedPlanRequest(name)
			res, err := queryClient.AppliedPlan(context.Background(), params)
			if err != nil {
				return err
			}

			if res.Height == 0 {
				return fmt.Errorf("no upgrade found")
			}

			// we got the height, now let's return the headers
			node, err := clientCtx.GetNode()
			if err != nil {
				return err
			}
			headers, err := node.BlockchainInfo(res.Height, res.Height)
			if err != nil {
				return err
			}
			if len(headers.BlockMetas) == 0 {
				return fmt.Errorf("no headers returned for height %d", res.Height)
			}

			// always output json as Header is unreable in toml ([]byte is a long list of numbers)
			bz, err := clientCtx.Codec.MarshalJSONIndent(headers.BlockMetas[0], "", "  ")
			if err != nil {
				return err
			}
			return clientCtx.PrintOutput(string(bz))
		},
	}
}
