package cli

import (
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/lib"
	wire "github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/ibc"
)

// flags
const (
	FlagSrcChainID    = "src-chain-id"
	FlagSrcChainNode  = "src-chain-node"
	FlagDestChainID   = "dest-chain-id"
	FlagDestChainNode = "dest-chain-node"
)

type relayCommander struct {
	cdc       *wire.Codec
	address   sdk.AccAddress
	decoder   auth.AccountDecoder
	mainStore string
	ibcStore  string
	accStore  string

	logger log.Logger
}

// IBC relay command
func IBCRelayCmd(cdc *wire.Codec) *cobra.Command {
	cmdr := relayCommander{
		cdc:       cdc,
		decoder:   authcmd.GetAccountDecoder(cdc),
		ibcStore:  "ibc",
		mainStore: "main",
		accStore:  "acc",

		logger: log.NewTMLogger(log.NewSyncWriter(os.Stdout)),
	}

	cmd := &cobra.Command{
		Use: "relay",
		Run: cmdr.runIBCRelay,
	}

	cmd.Flags().String(FlagSrcChainID, "", "Chain ID for ibc node to check outgoing packets")
	cmd.Flags().String(FlagSrcChainNode, "tcp://localhost:26657", "<host>:<port> to tendermint rpc interface for this chain")
	cmd.Flags().String(FlagDestChainID, "", "Chain ID for ibc node to broadcast incoming packets")
	cmd.Flags().String(FlagDestChainNode, "tcp://localhost:36657", "<host>:<port> to tendermint rpc interface for this chain")

	cmd.MarkFlagRequired(FlagSrcChainID)
	cmd.MarkFlagRequired(FlagSrcChainNode)
	cmd.MarkFlagRequired(FlagDestChainID)
	cmd.MarkFlagRequired(FlagDestChainNode)

	viper.BindPFlag(FlagSrcChainID, cmd.Flags().Lookup(FlagSrcChainID))
	viper.BindPFlag(FlagSrcChainNode, cmd.Flags().Lookup(FlagSrcChainNode))
	viper.BindPFlag(FlagDestChainID, cmd.Flags().Lookup(FlagDestChainID))
	viper.BindPFlag(FlagDestChainNode, cmd.Flags().Lookup(FlagDestChainNode))

	return cmd
}

// nolint: unparam
func (c relayCommander) runIBCRelay(cmd *cobra.Command, args []string) {
	srcChainID := viper.GetString(FlagSrcChainID)
	srcChainNode := viper.GetString(FlagSrcChainNode)
	toChainID := viper.GetString(FlagDestChainID)
	toChainNode := viper.GetString(FlagDestChainNode)
	address, err := context.NewCoreContextFromViper().GetFromAddress()
	if err != nil {
		panic(err)
	}
	c.address = address

	ctx := context.NewCoreContextFromViper()
	// TODO: use proper config
	egressQueue := lib.NewLinearClient(ctx.WithNodeURI(srcChainNode), "bank", c.cdc, []byte("ibc/"), nil)
	c.loop(egressQueue, srcChainID, toChainID, toChainNode)
}

func (c relayCommander) processed(node string, srcChainID string) uint64 {
	// TODO: Support receipts
	bz, err := query(node, ibc.IncomingSequenceKey(ibc.PacketType, srcChainID), c.ibcStore)
	if err != nil {
		panic(err)
	}

	if bz == nil {
		return 0
	}
	var res int64
	c.cdc.MustUnmarshalBinary(bz, &res)

	if res < 0 {
		panic("Negative processed result")
	}

	return uint64(res)
}

// This is nolinted as someone is in the process of refactoring this to remove the goto
// nolint: gocyclo
func (c relayCommander) loop(egressQueue lib.LinearClient, srcChainID, chainID, chainNode string) {
	ctx := context.NewCoreContextFromViper()

	for {
		time.Sleep(5 * time.Second)

		proc := c.processed(chainNode, srcChainID)
		length := egressQueue.Len()
		if length <= proc {
			continue
		}

		c.logger.Info("Detected IBC packet", "number", length-1)

		var data ibc.Datagram
		for i := proc; i < length; i++ {
			err := egressQueue.Get(i, &data)
			if err != nil {
				panic(err)
			}

			// TODO: add proof
			msg := ibc.MsgReceive{Datagram: data, Relayer: c.address}
			err = ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, []sdk.Msg{msg}, c.cdc)
			if err != nil {
				panic(err)
			}

			c.logger.Info("Relayed IBC packet", "number", i)
		}
	}
}

func query(node string, key []byte, storeName string) (res []byte, err error) {
	return context.NewCoreContextFromViper().WithNodeURI(node).QueryStore(key, storeName)
}

func (c relayCommander) broadcastTx(seq int64, node string, tx []byte) error {
	_, err := context.NewCoreContextFromViper().WithNodeURI(node).WithSequence(seq + 1).BroadcastTx(tx)
	return err
}

func (c relayCommander) getSequence(node string) int64 {
	res, err := query(node, c.address, c.accStore)
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
