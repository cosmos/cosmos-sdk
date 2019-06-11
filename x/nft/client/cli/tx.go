package cli

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/nft/types"
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

// GetCmdTransferNFT is the CLI command for sending a TransferNFT transaction
func GetCmdTransferNFT(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "transfer [sender] [recipient] [denom] [tokenID]",
		Short: "transfer a token of some denom with some tokenID to some recipient",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)
			txBldr := authtypes.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			tokenID := args[3]

			msg := types.NewMsgTransferNFT(sdk.AccAddress(args[0]), sdk.AccAddress(args[1]), args[2], tokenID)
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
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)
			txBldr := authtypes.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			denom := args[0]
			tokenID := args[1]

			name := viper.GetString(flagName)
			description := viper.GetString(flagDescription)
			image := viper.GetString(flagImage)
			tokenURI := viper.GetString(flagTokenURI)

			msg := types.NewMsgEditNFTMetadata(cliCtx.GetFromAddress(), denom, tokenID, name, description, image, tokenURI)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(flagName, "", "Name of the NFT")
	cmd.Flags().String(flagDescription, "", "Unique description of the NFT")
	cmd.Flags().String(flagImage, "", "Image path")
	cmd.Flags().String(flagTokenURI, "", "Extra properties available for querying")
	return cmd
}
