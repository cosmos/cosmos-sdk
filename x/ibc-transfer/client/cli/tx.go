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
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
)

// IBC transfer flags
var (
	FlagNode1    = "node1"
	FlagNode2    = "node2"
	FlagFrom1    = "from1"
	FlagFrom2    = "from2"
	FlagChainID2 = "chain-id2"
	FlagSequence = "packet-sequence"
	FlagTimeout  = "timeout"
)

// GetTransferTxCmd returns the command to create a NewMsgTransfer transaction
func GetTransferTxCmd(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "transfer [src-port] [src-channel] [dest-height] [receiver] [amount] [timeout-height] [timeout-timestamp]",
		Short: "Transfer fungible token through IBC",
		Args:  cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := authtypes.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc).WithBroadcastMode(flags.BroadcastBlock)

			sender := cliCtx.GetFromAddress()
			srcPort := args[0]
			srcChannel := args[1]

			destHeight, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid destination height: %w", err)
			}

			receiver := args[3]

			// parse coin trying to be sent
			coins, err := sdk.ParseCoins(args[4])
			if err != nil {
				return err
			}

			timeoutHeight, err := strconv.ParseUint(args[5], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid timeout height: %w", err)
			}

			timeoutTimestamp, err := strconv.ParseUint(args[6], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid timeout timestamp: %w", err)
			}

			msg := types.NewMsgTransfer(
				srcPort, srcChannel, destHeight, coins, sender, receiver,
				timeoutHeight, timeoutTimestamp,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
