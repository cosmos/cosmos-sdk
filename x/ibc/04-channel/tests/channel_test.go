package channel

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	tmclient "github.com/cosmos/cosmos-sdk/x/ibc/02-client/tendermint"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/tendermint/tests"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

func registerCodec(cdc *codec.Codec) {
	client.RegisterCodec(cdc)
	tmclient.RegisterCodec(cdc)
	commitment.RegisterCodec(cdc)
	merkle.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	channel.RegisterCodec(cdc)
	cdc.RegisterConcrete(MyPacket{}, "test/MyPacket", nil)
}

func TestHandshake(t *testing.T) {
	cdc := codec.New()
	registerCodec(cdc)

	node := NewNode(tendermint.NewMockValidators(100, 10), tendermint.NewMockValidators(100, 10), cdc)

	node.Handshake(t)
}

type MyPacket struct {
	Message string
}

func (packet MyPacket) Commit() []byte {
	return []byte(packet.Message)
}

func (packet MyPacket) Timeout() uint64 {
	return 100 // TODO
}

func TestPacket(t *testing.T) {
	cdc := codec.New()
	registerCodec(cdc)

	node := NewNode(tendermint.NewMockValidators(100, 10), tendermint.NewMockValidators(100, 10), cdc)

	node.Handshake(t)

	node.Send(t, MyPacket{"ping"})
	header := node.Commit()

	node.Counterparty.UpdateClient(t, header)
	cliobj := node.CLIObject()
	_, ppacket := node.Query(t, cliobj.PacketCommitKey(1))
	node.Counterparty.Receive(t, MyPacket{"ping"}, ppacket)
}
