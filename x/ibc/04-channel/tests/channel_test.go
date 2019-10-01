package channel

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	tmclient "github.com/cosmos/cosmos-sdk/x/ibc/02-client/tendermint"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/02-client/tendermint/tests"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
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

func (MyPacket) SenderPort() string {
	return PortName
}

func (MyPacket) ReceiverPort() string {
	return PortName
}

func (MyPacket) Type() string {
	return "my-packet"
}

func (MyPacket) ValidateBasic() sdk.Error {
	return nil
}

func (packet MyPacket) Marshal() []byte {
	cdc := codec.New()
	registerCodec(cdc)
	return cdc.MustMarshalBinaryBare(packet)
}

func TestPacket(t *testing.T) {
	cdc := codec.New()
	registerCodec(cdc)

	node := NewNode(tendermint.NewMockValidators(100, 10), tendermint.NewMockValidators(100, 10), cdc)

	node.Handshake(t)

	node.Send(t, MyPacket{"ping"})
	header := node.Commit()

	node.Counterparty.UpdateClient(t, header)
	cliobj := node.CLIState()
	_, ppacket := node.QueryValue(t, cliobj.Packets.Value(1))
	node.Counterparty.Receive(t, MyPacket{"ping"}, uint64(header.Height), ppacket)
}
