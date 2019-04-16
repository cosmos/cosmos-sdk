package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/nfts"
	"github.com/spf13/cobra"
)

// QueryBalanceOf = "balanceOf"
// QueryOwnerOf   = "ownerOf"
// QueryMetadata  = "metadata"

// GetCmdQueryBalanceOf queries balance of an account per some denom
func GetCmdQueryBalanceOf(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "balanceOf [denom] [accountAddress]",
		Short: "balanceOf denom accountAddress",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			denom := args[0]
			account := args[1]

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/balanceOf/%s/%s", queryRoute, denom, account), nil)
			if err != nil {
				fmt.Printf("could not query %s balance of account %s \n", denom, account)
				fmt.Print(err.Error())
				return nil
			}

			var out nfts.QueryResBalance
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdQueryOwnerOf queries owner of an nft per some denom
func GetCmdQueryOwnerOf(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "ownerOf [denom] [tokenID]",
		Short: "ownerOf denom tokenID",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			denom := args[0]
			tokenID := args[1]

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/ownerOf/%s/%s", queryRoute, denom, tokenID), nil)
			if err != nil {
				fmt.Printf("could not query owner of %s #%s\n", denom, tokenID)
				fmt.Print(err.Error())
				return nil
			}

			var out nfts.QueryResOwnerOf
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdQueryMetadata queries owner of an nft per some denom
func GetCmdQueryMetadata(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "metadata [denom] [tokenID]",
		Short: "metadata denom tokenID",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			denom := args[0]
			tokenID := args[1]

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/metadata/%s/%s", queryRoute, denom, tokenID), nil)
			if err != nil {
				fmt.Printf("could not query metadata of %s #%s\n", denom, tokenID)
				fmt.Print(err.Error())
				return nil
			}

			var out nfts.NFT
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}
