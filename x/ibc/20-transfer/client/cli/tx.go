package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// IBC transfer flags
var (
	FlagSource = "source"
)

// GetTxCmd returns the transaction commands for IBC fungible token transfer
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "transfer",
		Short: "IBC fungible token transfer transaction subcommands",
	}
	txCmd.AddCommand(
		GetTransferTxCmd(cdc),
	)

	return txCmd
}

// GetTransferTxCmd returns the command to create a NewMsgTransfer transaction
func GetTransferTxCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer [src-port] [src-channel] [receiver] [amount]",
		Short: "Transfer fungible token through IBC",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ctx := context.NewCLIContext().WithCodec(cdc).WithBroadcastMode(flags.BroadcastBlock)

			sender := ctx.GetFromAddress()
			srcPort := args[0]
			srcChan := args[1]
			receiver := args[2]

			// parse coin trying to be sent
			coin, err := sdk.ParseCoin(args[3])
			if err != nil {
				return err
			}

			source := false
			if viper.GetBool(FlagSource) {
				source = true
			}

			msg := types.NewMsgTransfer(srcPort, srcChan, coin.Denom, coin.Amount, sender, receiver, source)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(ctx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().Bool(FlagSource, false, "Pass flag for sending token from the source chain")

	cmd = client.PostCommands(cmd)[0]

	return cmd
}
