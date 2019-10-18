package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	mockbank "github.com/cosmos/cosmos-sdk/x/ibc/mock/bank"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "ibcmockbank",
		Short: "IBC mockbank module transaction subcommands",
		// RunE:  client.ValidateCmd,
	}
	txCmd.AddCommand(
		TransferTxCmd(cdc),
	)

	return txCmd
}

func TransferTxCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer --src-port <src port> --src-channel <src channel> --denom <denomination> --amount <amount> --receiver <receiver> --source <source>",
		Short: "Transfer tokens across chains through IBC",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ctx := context.NewCLIContext().WithCodec(cdc).WithBroadcastMode(flags.BroadcastBlock)

			sender := ctx.GetFromAddress()
			receiver := viper.GetString(FlagReceiver)
			denom := viper.GetString(FlagDenom)
			srcPort := viper.GetString(FlagSrcPort)
			srcChan := viper.GetString(FlagSrcChannel)
			source := viper.GetBool(FlagSource)

			amount, ok := sdk.NewIntFromString(viper.GetString(FlagAmount))
			if !ok {
				return fmt.Errorf("invalid amount")
			}

			msg := mockbank.NewMsgTransfer(srcPort, srcChan, denom, amount, sender, receiver, source)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(ctx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.MarkFlagRequired(FlagSrcPort)
	cmd.MarkFlagRequired(FlagSrcChannel)
	cmd.MarkFlagRequired(FlagDenom)
	cmd.MarkFlagRequired(FlagAmount)
	cmd.MarkFlagRequired(FlagReceiver)

	cmd = client.PostCommands(cmd)[0]

	return cmd
}
