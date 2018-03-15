package commands

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/builder"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"

	"github.com/cosmos/cosmos-sdk/x/ibc"
)

const (
	flagChain1 = "chain1"
	flagChain2 = "chain2"
)

func IBCRelayCmd(cdc *wire.Codec) *cobra.Command {
	cmdr := relayCommander{
		cdc:      cdc,
		ibcStore: "ibc",
	}

	cmd := &cobra.Command{
		Use: "relay",
		Run: cmdr.runIBCRelay,
	}
	cmd.Flags().String(flagChain1, "", "Chain ID to relay IBC packets")
	cmd.Flags().String(flagChain2, "", "Chain ID to relay IBC packets")
	return cmd
}

type relayCommander struct {
	cdc      *wire.Codec
	address  sdk.Address
	ibcStore string
}

func (c relayCommander) runIBCRelay(cmd *cobra.Command, args []string) {
	chain1 := viper.GetString(flagChain1)
	chain2 := viper.GetString(flagChain2)

	address, err := builder.GetFromAddress()
	if err != nil {
		panic(err)
	}
	c.address = address

	go c.loop(chain1, chain2)
	go c.loop(chain2, chain1)
}

// https://github.com/cosmos/cosmos-sdk/blob/master/client/helpers.go using specified address

func query(id string, key []byte, storeName string) (res []byte, err error) {
	orig := viper.GetString(client.FlagNode)
	viper.Set(client.FlagNode, id)
	res, err = builder.Query(key, storeName)
	viper.Set(client.FlagNode, orig)
	return res, err
}

func broadcastTx(id string, tx []byte) error {
	orig := viper.GetString(client.FlagNode)
	viper.Set(client.FlagNode, id)
	_, err := builder.BroadcastTx(tx)
	viper.Set(client.FlagNode, orig)
	return err
}

func (c relayCommander) refine(bz []byte, sequence int64) []byte {
	var packet ibc.IBCPacket
	if err := c.cdc.UnmarshalBinary(bz, &packet); err != nil {
		panic(err)
	}

	msg := ibc.IBCReceiveMsg{
		IBCPacket: packet,
		Relayer:   c.address,
		Sequence:  sequence,
	}
	res, err := builder.SignAndBuild(msg, c.cdc)
	if err != nil {
		panic(err)
	}
	return res
}

func (c relayCommander) loop(fromID, toID string) {
	ingressKey := ibc.IngressKey(fromID)

	processedbz, err := query(toID, ingressKey, c.ibcStore)
	if err != nil {
		panic(err)
	}

	var processed int64
	if err = c.cdc.UnmarshalBinary(processedbz, &processed); err != nil {
		panic(err)
	}

OUTER:
	for {
		time.Sleep(time.Second)

		lengthKey := ibc.EgressLengthKey(toID)
		egressLengthbz, err := query(fromID, lengthKey, c.ibcStore)
		if err != nil {
			fmt.Printf("Error querying outgoing packet list length: '%s'\n", err)
			continue OUTER
		}
		var egressLength int64
		if err = c.cdc.UnmarshalBinary(egressLengthbz, &egressLength); err != nil {
			panic(err)
		}

		for i := processed; i < egressLength; i++ {
			egressbz, err := query(fromID, ibc.EgressKey(toID, i), c.ibcStore)
			if err != nil {
				fmt.Printf("Error querying egress packet: '%s'\n", err)
				continue OUTER
			}

			err = broadcastTx(toID, c.refine(egressbz, i))
			if err != nil {
				fmt.Printf("Error broadcasting ingress packet: '%s'\n", err)
				continue OUTER
			}

			fmt.Printf("Relayed packet: %d\n", i)
		}

		processed = egressLength
	}
}
