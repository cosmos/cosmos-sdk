package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/nft/internal/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Edit metadata flags
const (
	flagName        = "name"
	flagDescription = "description"
	flagImage       = "image"
	flagTokenURI    = "tokenURI"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	nftTxCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "NFT transactions subcommands",
	}

	nftTxCmd.AddCommand(client.PostCommands(
		GetCmdTransferNFT(cdc),
		GetCmdEditNFTMetadata(cdc),
		GetCmdMintNFT(cdc),
		GetCmdBurnNFT(cdc),
	)...)

	return nftTxCmd
}

// GetCmdTransferNFT is the CLI command for sending a TransferNFT transaction
func GetCmdTransferNFT(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "transfer [sender] [recipient] [denom] [tokenID]",
		Short: "transfer a token of some denom with some tokenID to some recipient",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := authtypes.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			sender, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			recipient, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			denom := args[2]
			tokenID := args[3]

			msg := types.NewMsgTransferNFT(sender, recipient, denom, tokenID)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdEditNFTMetadata is the CLI command for sending an EditMetadata transaction
func GetCmdEditNFTMetadata(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit-metadata [denom] [tokenID]",
		Short: "transfer a token of some denom with some tokenID to some recipient",
		Args:  cobra.ExactArgs(2),

		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := authtypes.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			denom := args[0]
			tokenID := args[1]

			name := viper.GetString(flagName)
			description := viper.GetString(flagDescription)
			image := viper.GetString(flagImage)
			tokenURI := viper.GetString(flagTokenURI)

			msg := types.NewMsgEditNFTMetadata(cliCtx.GetFromAddress(), tokenID, denom, name, description, image, tokenURI)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(flagName, "", "Name of the NFT")
	cmd.Flags().String(flagDescription, "", "Unique description of the NFT")
	cmd.Flags().String(flagImage, "", "Image path")
	cmd.Flags().String(flagTokenURI, "", "Extra properties available for querying")
	return cmd
}

// GetCmdMintNFT is the CLI command for a MintNFT transaction
func GetCmdMintNFT(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint [recipient] [denom] [tokenID]",
		Short: "mints a token of some denom with some tokenID to some recipient",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := authtypes.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			denom := args[1]
			tokenID := args[2]

			recipient, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			name := viper.GetString(flagName)
			description := viper.GetString(flagDescription)
			image := viper.GetString(flagImage)
			tokenURI := viper.GetString(flagTokenURI)

			msg := types.NewMsgMintNFT(cliCtx.GetFromAddress(), recipient, tokenID, denom, name, description, image, tokenURI)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(flagName, "", "Name/title of nft")
	cmd.Flags().String(flagDescription, "", "Description of nft")
	cmd.Flags().String(flagImage, "", "Image uri of nft")
	cmd.Flags().String(flagTokenURI, "", "URI for supplemental off-chain metadata (should return a JSON object)")

	return cmd
}

// GetCmdBurnNFT is the CLI command for sending a BurnNFT transaction
func GetCmdBurnNFT(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "burn [denom] [tokenID]",
		Short: "burn a token of some denom with some tokenID",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := authtypes.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			denom := args[0]
			tokenID := args[1]

			msg := types.NewMsgBurnNFT(cliCtx.GetFromAddress(), tokenID, denom)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
