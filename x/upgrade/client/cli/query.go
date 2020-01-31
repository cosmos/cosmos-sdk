package cli

import (
	"encoding/binary"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/internal/types"
)

// GetPlanCmd returns the query upgrade plan command
func GetPlanCmd(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "plan",
		Short: "get upgrade plan (if one exists)",
		Long:  "Gets the currently scheduled upgrade plan, if one exists",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// ignore height for now
			res, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s", upgrade.QuerierKey, upgrade.QueryCurrent))
			if err != nil {
				return err
			}

			if len(res) == 0 {
				return fmt.Errorf("no upgrade scheduled")
			}

			var plan upgrade.Plan
			err = cdc.UnmarshalJSON(res, &plan)
			if err != nil {
				return err
			}
			return cliCtx.PrintOutput(plan)
		},
	}
}

// GetAppliedHeightCmd returns the height at which a completed upgrade was applied
func GetAppliedHeightCmd(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "applied [upgrade-name]",
		Short: "block header for height at which a completed upgrade was applied",
		Long: "If upgrade-name was previously executed on the chain, this returns the header for the block at which it was applied.\n" +
			"This helps a client determine which binary was valid over a given range of blocks, as well as more context to understand past migrations.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			name := args[0]
			params := upgrade.NewQueryAppliedParams(name)
			bz, err := cliCtx.Codec.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", upgrade.QuerierKey, upgrade.QueryApplied), bz)
			if err != nil {
				return err
			}

			if len(res) == 0 {
				return fmt.Errorf("no upgrade found")
			}
			if len(res) != 8 {
				return fmt.Errorf("unknown format for applied-upgrade")
			}
			applied := int64(binary.BigEndian.Uint64(res))

			// we got the height, now let's return the headers
			node, err := cliCtx.GetNode()
			if err != nil {
				return err
			}
			headers, err := node.BlockchainInfo(applied, applied)
			if err != nil {
				return err
			}
			if len(headers.BlockMetas) == 0 {
				return fmt.Errorf("no headers returned for height %d", applied)
			}

			// always output json as Header is unreable in toml ([]byte is a long list of numbers)
			bz, err = cdc.MarshalJSONIndent(headers.BlockMetas[0], "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(bz))
			return nil
		},
	}
}
