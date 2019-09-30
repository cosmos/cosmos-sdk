package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/nft/exported"
	"github.com/cosmos/cosmos-sdk/x/nft/internal/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	nftQueryCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Querying commands for the NFT module",
	}

	nftQueryCmd.AddCommand(client.GetCommands(
		GetCmdQueryCollectionSupply(queryRoute, cdc),
		GetCmdQueryOwner(queryRoute, cdc),
		GetCmdQueryCollection(queryRoute, cdc),
		GetCmdQueryDenoms(queryRoute, cdc),
		GetCmdQueryNFT(queryRoute, cdc),
	)...)

	return nftQueryCmd
}

// GetCmdQueryCollectionSupply queries the supply of a nft collection
func GetCmdQueryCollectionSupply(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "supply [denom]",
		Short: "total supply of a collection of NFTs",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Get the total count of NFTs that match a certain denomination.

Example:
$ %s query %s supply crypto-kitties
`, version.ClientName, types.ModuleName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			denom := args[0]

			params := types.NewQueryCollectionParams(denom)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/supply/%s", queryRoute, denom), bz)
			if err != nil {
				return err
			}

			var out exported.NFT
			err = cdc.UnmarshalJSON(res, &out)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdQueryOwner queries all the NFTs owned by an account
func GetCmdQueryOwner(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "owner [accountAddress] [denom]",
		Short: "get the NFTs owned by an account address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Get the NFTs owned by an account address optionally filtered by the denom of the NFTs.

Example:
$ %s query %s owner cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
$ %s query %s owner cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p cripto-kitties
`, version.ClientName, types.ModuleName, version.ClientName, types.ModuleName,
			),
		),
		Args: cobra.RangeArgs(1, 2),
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
				res, _, err = cliCtx.QueryWithData(fmt.Sprintf("custom/%s/owner", queryRoute), bz)
			} else {
				res, _, err = cliCtx.QueryWithData(fmt.Sprintf("custom/%s/ownerByDenom", queryRoute), bz)
			}

			if err != nil {
				return err
			}

			var out types.Owner
			err = cdc.UnmarshalJSON(res, &out)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdQueryCollection queries all the NFTs from a collection
func GetCmdQueryCollection(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "collection [denom]",
		Short: "get all the NFTs from a given collection",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Get a list of all NFTs from a given collection.

Example:
$ %s query %s collection cripto-kitties
`, version.ClientName, types.ModuleName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			denom := args[0]

			params := types.NewQueryCollectionParams(denom)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/collection", queryRoute), bz)
			if err != nil {
				return err
			}

			var out types.Collection
			err = cdc.UnmarshalJSON(res, &out)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdQueryDenoms queries all denoms
func GetCmdQueryDenoms(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "denoms",
		Short: "queries all denominations of all collections of NFTs",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Gets all denominations of all the available collections of NFTs that
			are stored on the chain.

			Example:
			$ %s query %s denoms
			`, version.ClientName, types.ModuleName,
			),
		),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/denoms", queryRoute), nil)
			if err != nil {
				return err
			}

			var out types.SortedStringArray
			err = cdc.UnmarshalJSON(res, &out)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdQueryNFT queries a single NFTs from a collection
func GetCmdQueryNFT(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "token [denom] [ID]",
		Short: "query a single NFT from a collection",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Get an NFT from a collection that has the given ID (SHA-256 hex hash).

Example:
$ %s query %s token crypto-kitties d04b98f48e8f8bcc15c6ae5ac050801cd6dcfd428fb5f9e65c4e16e7807340fa
`, version.ClientName, types.ModuleName,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			denom := args[0]
			id := args[1]

			params := types.NewQueryNFTParams(denom, id)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/nft", queryRoute), bz)
			if err != nil {
				return err
			}

			var out exported.NFT
			err = cdc.UnmarshalJSON(res, &out)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(out)
		},
	}
}
