package cli

import (
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
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
		Use:   "transfer <src-port> <src-channel> <amount> <receiver> <source> <timeout> [proof-path] [proof-height]",
		Short: "Transfer",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ctx := context.NewCLIContext().WithCodec(cdc).WithBroadcastMode(flags.BroadcastBlock)

			sender := ctx.GetFromAddress()
			srcPort := viper.GetString(FlagSrcPort)
			srcChan := viper.GetString(FlagSrcChannel)
			source := viper.GetBool(FlagSource)
			timeout := viper.GetUint64(FlagTimeout)

			amountStr := viper.GetString(FlagAmount)
			amount, err := sdk.ParseCoin(amountStr)
			if err != nil {
				return err
			}

			receiver, err := sdk.AccAddressFromBech32(viper.GetString(FlagReceiver))
			if err != nil {
				return err
			}

			var proof ics23.Proof
			var proofHeight uint64

			proofPath := viper.GetString(FlagProofPath)
			if proofPath != "" {
				proofBz, err := ioutil.ReadFile(proofPath)
				if err != nil {
					return err
				}

				if err := cdc.UnmarshalJSON(proofBz, &proof); err != nil {
					return err
				}

				proofHeight = viper.GetUint64(FlagProofHeight)
			}

			msg := mockbank.NewMsgTransfer(srcPort, srcChan, amount, sender, receiver, source, timeout, proof, proofHeight)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(ctx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd = client.PostCommands(cmd)[0]

	return cmd
}
