package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"

	"github.com/cosmos/cosmos-sdk/x/ibc"
)

const (
	flagTo     = "to"
	flagAmount = "amount"
	flagChain  = "chain"
)

func IBCTransferCmd(cdc *wire.Codec) *cobra.Command {
	cmdr := sendCommander{cdc}
	cmd := &cobra.Command{
		Use:  "transfer",
		RunE: cmdr.sendIBCTransfer,
	}
	cmd.Flags().String(flagTo, "", "Address to send coins")
	cmd.Flags().String(flagAmount, "", "Amount of coins to send")
	cmd.Flags().String(flagChain, "", "Destination chain to send coins")
	return cmd
}

type sendCommander struct {
	cdc *wire.Codec
}

func (c sendCommander) sendIBCTransfer(cmd *cobra.Command, args []string) error {
	ctx := context.NewCoreContextFromViper()

	// get the from address
	from, err := ctx.GetFromAddress()
	if err != nil {
		return err
	}

	// build the message
	msg, err := buildMsg(from)
	if err != nil {
		return err
	}

	// get password
	res, err := ctx.SignBuildBroadcast(ctx.FromAddressName, msg, c.cdc)
	if err != nil {
		return err
	}

	fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
	return nil
}

func buildMsg(from sdk.Address) (sdk.Msg, error) {
	amount := viper.GetString(flagAmount)
	coins, err := sdk.ParseCoins(amount)
	if err != nil {
		return nil, err
	}

	dest := viper.GetString(flagTo)
	bz, err := hex.DecodeString(dest)
	if err != nil {
		return nil, err
	}
	to := sdk.Address(bz)

	packet := ibc.NewIBCPacket(from, to, coins, viper.GetString(client.FlagChainID),
		viper.GetString(flagChain))

	msg := ibc.IBCTransferMsg{
		IBCPacket: packet,
	}

	return msg, nil
}
