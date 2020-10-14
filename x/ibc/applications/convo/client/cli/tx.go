package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/ibc/applications/convo/types"
	channelutils "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/client/utils"
)

// NewConvoTxCmd returns the command to create a NewConvoTx transaction
func NewConvoTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "convo [src-port] [src-channel] [receiver] [message]",
		Short:   "Send a message to receiver over IBC",
		Long:    "Send a message to receiver over IBC",
		Example: fmt.Sprintf("%s tx ibc-convo convo [src-port] [src-channel] [receiver] [message]", version.AppName),
		Args:    cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			sender := clientCtx.GetFromAddress()
			srcPort := args[0]
			srcChannel := args[1]
			receiver := args[2]
			message := args[3]

			// if the timeouts are not absolute, retrieve latest block height and block timestamp
			// for the consensus state connected to the destination port/channel
			_, height, _, err := channelutils.QueryLatestConsensusState(clientCtx, srcPort, srcChannel)
			if err != nil {
				return err
			}

			// Make timeout height 10000 blocks past current height
			timeoutHeight := height
			timeoutHeight.VersionHeight += 10000

			msg := types.NewMsgConvo(
				srcPort, srcChannel, sender, receiver, message, timeoutHeight, 0,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
