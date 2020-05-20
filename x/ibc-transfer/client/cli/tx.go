package cli

import (
	"bufio"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
)

// GetTransferTxCmd returns the command to create a NewMsgTransfer transaction
func GetTransferTxCmd(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "transfer [src-port] [src-channel] [receiver] [amount] [timeout-height] [timeout-timestamp]",
		Short:   "Transfer a fungible token through IBC",
		Example: fmt.Sprintf("%s tx transfer transfer [src-port] [src-channel] [receiver] [amount] [timeout-height] [timeout-timestamp]", version.ClientName),
		Args:    cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := authtypes.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc).WithBroadcastMode(flags.BroadcastBlock)

			sender := cliCtx.GetFromAddress()
			srcPort := args[0]
			srcChannel := args[1]
			receiver := args[2]

			coins, err := sdk.ParseCoins(args[3])
			if err != nil {
				return err
			}

			timeoutHeight, err := strconv.ParseUint(args[4], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid timeout height: %w", err)
			}

			timeoutTimestamp, err := strconv.ParseUint(args[5], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid timeout timestamp: %w", err)
			}

			msg := types.NewMsgTransfer(
				srcPort, srcChannel, coins, sender, receiver, timeoutHeight, timeoutTimestamp,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
