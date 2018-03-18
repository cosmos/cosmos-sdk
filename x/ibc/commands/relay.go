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
	FlagFromChainID   = "from-chain-id"
	FlagFromChainNode = "from-chain-node"
	FlagToChainID     = "to-chain-id"
	FlagToChainNode   = "to-chain-node"
)

type relayCommander struct {
	cdc      *wire.Codec
	address  sdk.Address
	ibcStore string
}

func IBCRelayCmd(cdc *wire.Codec) *cobra.Command {
	cmdr := relayCommander{
		cdc:      cdc,
		ibcStore: "ibc",
	}

	cmd := &cobra.Command{
		Use: "relay",
		Run: cmdr.runIBCRelay,
	}

	cmd.Flags().String(client.FlagName, "", "Name of private key with which to sign")
	cmd.Flags().Bool(client.FlagTrustNode, true, "Don't verify proofs for responses")
	cmd.Flags().String(FlagFromChainID, "", "Chain ID for ibc node to check outgoing packets")
	cmd.Flags().String(FlagFromChainNode, "tcp://localhost:46657", "<host>:<port> to tendermint rpc interface for this chain")
	cmd.Flags().String(FlagToChainID, "", "Chain ID for ibc node to broadcast incoming packets")
	cmd.Flags().String(FlagToChainNode, "tcp://localhost:46658", "<host>:<port> to tendermint rpc interface for this chain")

	cmd.MarkFlagRequired(client.FlagName)
	cmd.MarkFlagRequired(FlagFromChainID)
	cmd.MarkFlagRequired(FlagFromChainNode)
	cmd.MarkFlagRequired(FlagToChainID)
	cmd.MarkFlagRequired(FlagToChainNode)

	viper.BindPFlag(client.FlagName, cmd.Flags().Lookup(client.FlagName))
	viper.BindPFlag(client.FlagTrustNode, cmd.Flags().Lookup(client.FlagTrustNode))
	viper.BindPFlag(FlagFromChainID, cmd.Flags().Lookup(FlagFromChainID))
	viper.BindPFlag(FlagFromChainNode, cmd.Flags().Lookup(FlagFromChainNode))
	viper.BindPFlag(FlagToChainID, cmd.Flags().Lookup(FlagToChainID))
	viper.BindPFlag(FlagToChainNode, cmd.Flags().Lookup(FlagToChainNode))

	return cmd
}

func (c relayCommander) runIBCRelay(cmd *cobra.Command, args []string) {
	fromChainID := viper.GetString(FlagFromChainID)
	fromChainNode := viper.GetString(FlagFromChainNode)
	toChainID := viper.GetString(FlagToChainID)
	toChainNode := viper.GetString(FlagToChainNode)
	address, err := builder.GetFromAddress()

	if err != nil {
		panic(err)
	}
	c.address = address

	c.loop(fromChainID, fromChainNode, toChainID, toChainNode)
}

func (c relayCommander) loop(fromChainID, fromChainNode, toChainID, toChainNode string) {
	ingressKey := ibc.IngressSequenceKey(fromChainID)

	processedbz, err := query(toChainNode, ingressKey, c.ibcStore)
	if err != nil {
		panic(err)
	}

	var processed int64
	if processedbz == nil {
		processed = 0
	} else if err = c.cdc.UnmarshalBinary(processedbz, &processed); err != nil {
		panic(err)
	}

OUTER:
	for {
		time.Sleep(time.Second)

		lengthKey := ibc.EgressLengthKey(toChainID)
		egressLengthbz, err := query(fromChainNode, lengthKey, c.ibcStore)
		if err != nil {
			fmt.Printf("Error querying outgoing packet list length: '%s'\n", err)
			continue OUTER
		}
		var egressLength int64
		if egressLengthbz == nil {
			egressLength = 0
		} else if err = c.cdc.UnmarshalBinary(egressLengthbz, &egressLength); err != nil {
			panic(err)
		}

		for i := processed; i < egressLength; i++ {
			egressbz, err := query(fromChainNode, ibc.EgressKey(toChainID, i), c.ibcStore)
			if err != nil {
				fmt.Printf("Error querying egress packet: '%s'\n", err)
				continue OUTER
			}

			err = broadcastTx(toChainNode, c.refine(egressbz, i))
			if err != nil {
				fmt.Printf("Error broadcasting ingress packet: '%s'\n", err)
				continue OUTER
			}

			fmt.Printf("Relayed packet: %d\n", i)
		}

		processed = egressLength
	}
}

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
