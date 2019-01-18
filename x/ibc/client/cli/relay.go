package cli

import (
	"os"
	"time"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/ibc"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/libs/log"
)

// flags
const (
	FlagFromChainID   = "from-chain-id"
	FlagFromChainNode = "from-chain-node"
	FlagToChainID     = "to-chain-id"
	FlagToChainNode   = "to-chain-node"
)

type relayCommander struct {
	cdc       *codec.Codec
	address   sdk.AccAddress
	decoder   auth.AccountDecoder
	mainStore string
	ibcStore  string
	accStore  string

	logger log.Logger
}

// IBCRelayCmd implements the IBC relay command.
func IBCRelayCmd(cdc *codec.Codec) *cobra.Command {
	cliCtx := context.NewCLIContext(cdc)
	cmdr := relayCommander{
		cdc:       cdc,
		decoder:   cliCtx.GetAccountDecoder(),
		ibcStore:  "ibc",
		mainStore: bam.MainStoreKey,
		accStore:  auth.StoreKey,

		logger: log.NewTMLogger(log.NewSyncWriter(os.Stdout)),
	}

	cmd := &cobra.Command{
		Use: "relay",
		Run: cmdr.runIBCRelay,
	}

	cmd.Flags().String(FlagFromChainID, "", "Chain ID for ibc node to check outgoing packets")
	cmd.Flags().String(FlagFromChainNode, "tcp://localhost:26657", "<host>:<port> to tendermint rpc interface for this chain")
	cmd.Flags().String(FlagToChainID, "", "Chain ID for ibc node to broadcast incoming packets")
	cmd.Flags().String(FlagToChainNode, "tcp://localhost:36657", "<host>:<port> to tendermint rpc interface for this chain")

	cmd.MarkFlagRequired(FlagFromChainID)
	cmd.MarkFlagRequired(FlagFromChainNode)
	cmd.MarkFlagRequired(FlagToChainID)
	cmd.MarkFlagRequired(FlagToChainNode)

	viper.BindPFlag(FlagFromChainID, cmd.Flags().Lookup(FlagFromChainID))
	viper.BindPFlag(FlagFromChainNode, cmd.Flags().Lookup(FlagFromChainNode))
	viper.BindPFlag(FlagToChainID, cmd.Flags().Lookup(FlagToChainID))
	viper.BindPFlag(FlagToChainNode, cmd.Flags().Lookup(FlagToChainNode))

	return cmd
}

// nolint: unparam
func (c relayCommander) runIBCRelay(cmd *cobra.Command, args []string) {
	fromChainID := viper.GetString(FlagFromChainID)
	fromChainNode := viper.GetString(FlagFromChainNode)
	toChainID := viper.GetString(FlagToChainID)
	toChainNode := viper.GetString(FlagToChainNode)

	c.address = context.NewCLIContext(c.cdc).FromAddr()
	c.loop(fromChainID, fromChainNode, toChainID, toChainNode)
}

// This is nolinted as someone is in the process of refactoring this to remove the goto
func (c relayCommander) loop(fromChainID, fromChainNode, toChainID, toChainNode string) {
	cliCtx := context.NewCLIContext(c.cdc)

	passphrase, err := keys.ReadPassphraseFromStdin(cliCtx.FromName())
	if err != nil {
		panic(err)
	}

	ingressKey := ibc.IngressSequenceKey(fromChainID)
	lengthKey := ibc.EgressLengthKey(toChainID)

OUTER:
	for {
		time.Sleep(5 * time.Second)

		processedbz, err := c.query(toChainNode, ingressKey, c.ibcStore)
		if err != nil {
			panic(err)
		}

		var processed uint64
		if processedbz == nil {
			processed = 0
		} else if err = c.cdc.UnmarshalBinaryLengthPrefixed(processedbz, &processed); err != nil {
			panic(err)
		}

		egressLengthbz, err := c.query(fromChainNode, lengthKey, c.ibcStore)
		if err != nil {
			c.logger.Error("error querying outgoing packet list length", "err", err)
			continue OUTER // TODO replace with continue (I think it should just to the correct place where OUTER is now)
		}

		var egressLength uint64
		if egressLengthbz == nil {
			egressLength = 0
		} else if err = c.cdc.UnmarshalBinaryLengthPrefixed(egressLengthbz, &egressLength); err != nil {
			panic(err)
		}

		if egressLength > processed {
			c.logger.Info("Detected IBC packet", "number", egressLength-1)
		}

		seq := c.getSequence(toChainNode)

		for i := processed; i < egressLength; i++ {
			egressbz, err := c.query(fromChainNode, ibc.EgressKey(toChainID, i), c.ibcStore)
			if err != nil {
				c.logger.Error("error querying egress packet", "err", err)
				continue OUTER // TODO replace to break, will break first loop then send back to the beginning (aka OUTER)
			}

			err = c.broadcastTx(toChainNode, c.refine(egressbz, i, seq, passphrase))

			seq++

			if err != nil {
				c.logger.Error("error broadcasting ingress packet", "err", err)
				continue OUTER // TODO replace to break, will break first loop then send back to the beginning (aka OUTER)
			}

			c.logger.Info("Relayed IBC packet", "number", i)
		}
	}
}

func (c relayCommander) query(node string, key []byte, storeName string) (res []byte, err error) {
	return context.NewCLIContext(c.cdc).SetNode(node).QueryStore(key, storeName)
}

// nolint: unparam
func (c relayCommander) broadcastTx(node string, tx []byte) error {
	_, err := context.NewCLIContext(c.cdc).SetNode(node).BroadcastTx(tx)
	return err
}

func (c relayCommander) getSequence(node string) uint64 {
	res, err := c.query(node, auth.AddressStoreKey(c.address), c.accStore)
	if err != nil {
		panic(err)
	}

	if nil != res {
		account, err := c.decoder(res)
		if err != nil {
			panic(err)
		}

		return account.GetSequence()
	}

	return 0
}

func (c relayCommander) refine(bz []byte, ibcSeq, accSeq uint64, passphrase string) []byte {
	var packet ibc.IBCPacket
	if err := c.cdc.UnmarshalBinaryLengthPrefixed(bz, &packet); err != nil {
		panic(err)
	}

	msg := ibc.IBCReceiveMsg{
		IBCPacket: packet,
		Relayer:   c.address,
		Sequence:  ibcSeq,
	}

	cliCtx := context.NewCLIContextTx(c.cdc)

	res, err := cliCtx.TxBldr.BuildAndSign(cliCtx.FromName(), passphrase, []sdk.Msg{msg})
	if err != nil {
		panic(err)
	}

	return res
}
