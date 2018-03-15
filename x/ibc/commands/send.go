package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/builder"

	sdk "github.com/cosmos/cosmos-sdk/types"
	//	wire "github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/x/ibc"
)

func IBCTransferCmd(cdc *wire.Codec) *cobra.Command {
	cmdr := sendCommander{cdc}

	cmd := &cobra.Command{
		Use:  "transfer",
		RunE: cmdr.runIBCTransfer,
	}
	cmd.Flags().String(flagTo, "", "Address to send coins")
	cmd.Flags().String(flagAmount, "", "Amount of coins to send")
	cmd.Flags().String(flagChain, "", "Destination chain to send coins")
	viper.BindPFlag(flagTo, cmd.Flags().Lookup(flagTo))
	viper.BindPFlag(flagAmount, cmd.Flags().Lookup(flagAmount))
	viper.BindPFlag(flagChain, cmd.Flags().Lookup(flagChain))
	cmd.MarkFlagRequired(flagTo)
	cmd.MarkFlagRequired(flagAmount)
	cmd.MarkFlagRequired(flagChain)

	return cmd
}

type sendCommander struct {
	cdc *wire.Codec
}

func (c sendCommander) runIBCTransfer(cmd *cobra.Command, args []string) error {
	keyname := viper.GetString(client.FlagName)

	address := getAddress(keyname)
	msg, err := buildMsg(address)
	if err != nil {
		return err
	}

	txBytes, err := buildTx(c.cdc, msg, keyname)
	if err != nil {
		return err
	}

	res, err := builder.BroadcastTx(txBytes)
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

	return ibc.IBCTransferMsg{
		IBCPacket: ibc.IBCPacket{
			SrcAddr:   from,
			DestAddr:  to,
			Coins:     coins,
			SrcChain:  viper.GetString(client.FlagNode),
			DestChain: viper.GetString(flagChain),
		},
	}, nil
}
