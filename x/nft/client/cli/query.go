package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft/types"
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

			params := types.NewQueryCollectionParams(denom)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/supply/%s", queryRoute, denom), bz)
			if err != nil {
				return err
			}

			var out types.NFT
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdQueryOwner queries all the NFTs owned by an account
func GetCmdQueryOwner(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "owner [accountAddress] [denom (optional)]",
		Short: "get the NFTs owned by an account address",
		Long:  "get the NFTs owned by an account address optionally filtered by the denom of the NFTs",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			address, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			denom := ""
			if len(args) == 2 {
				denom = args[1]
			}

			params := types.NewQueryBalanceParams(address, denom)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			var res []byte
			if denom == "" {
				res, err = cliCtx.QueryWithData(fmt.Sprintf("custom/%s/owner", queryRoute), bz)
			} else {
				res, err = cliCtx.QueryWithData(fmt.Sprintf("custom/%s/ownerByDenom", queryRoute), bz)
			}

			if err != nil {
				return err
			}

			var out types.Owner
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdQueryCollection queries all the NFTs from a collection
func GetCmdQueryCollection(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "collection [denom]",
		Short: "get all the NFTs from a given collection",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			denom := args[0]

			params := types.NewQueryCollectionParams(denom)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/collection", queryRoute), bz)
			if err != nil {
				return err
			}

			var out types.Collection
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdQueryNFT queries a single NFTs from a collection
func GetCmdQueryNFT(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "token [denom] [ID]",
		Short: "query a single NFT from a collection",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			denom := args[0]
			id := args[1]

			params := types.NewQueryNFTParams(denom, id)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/nft", queryRoute), bz)
			if err != nil {
				return err
			}

			var out types.NFT
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}
