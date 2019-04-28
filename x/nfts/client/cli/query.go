package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/nfts"
	"github.com/spf13/cobra"
)

// GetCmdQueryCollectionSupply queries the supply of a nft collection
func GetCmdQueryCollectionSupply(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "supply [denom]",
		Short: "total supply of a collection of NFTs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			denom := args[0]

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/supply/%s", queryRoute, denom), nil)
			if err != nil {
				fmt.Printf("could not query supply of %s\n", denom)
				fmt.Print(err.Error())
				return nil
			}

			// var out nfts.NFT
			// cdc.MustUnmarshalJSON(res, &out)
			// return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdQueryBalance queries all the NFTs owned by an account
func GetCmdQueryBalance(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "balance [accountAddress] [denom]",
		Short: "get the NFTs owned by an account address",
		Long:  "get the NFTs owned by an account address", // TODO: finish this
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			account := args[0]

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/balance/%s", queryRoute, account), nil)
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

// GetCmdQueryNFTs queries all the NFTs from a collection
func GetCmdQueryNFTs(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "collection [denom]",
		Short: "get all the NFTs from a given collection",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			denom := args[0]
			tokenID := args[1]

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/collection/%s", queryRoute, denom, tokenID), nil)
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

// GetCmdQueryNFT queries a single NFTs from a collection
func GetCmdQueryNFT(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "nft [denom] [ID]",
		Short: "query a single NFT from a collection",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			denom := args[0]
			ID := args[1]

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/nft/%s/%s", queryRoute, denom, tokenID), nil)
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
