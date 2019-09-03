package channel

import (
	"errors"
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

func (packet MyPacket) MarshalAmino() (string, error) {
	return "mp-" + packet.Message, nil
}

func (packet *MyPacket) UnmarshalAmino(text string) error {
	if text[:3] != "mp-" {
		return errors.New("Invalid text for MyPacket")
	}
	packet.Message = text[3:]
	return nil
}

func (packet MyPacket) MarshalJSON() ([]byte, error) {
	res, _ := packet.MarshalAmino()
	return []byte("\"" + res + "\""), nil
}

func (packet *MyPacket) UnmarshalJSON(bz []byte) error {
	bz = bz[1 : len(bz)-1]
	return packet.UnmarshalAmino(string(bz))
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
	_, ppacket := node.QueryValue(t, cliobj.Packets.Value(1))
	node.Counterparty.Receive(t, MyPacket{"ping"}, ppacket)
}
