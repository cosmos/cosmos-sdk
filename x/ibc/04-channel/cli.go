package channel

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

// CLIObject stores the key for each object fields
type CLIObject struct {
	ChanID     string
	ChannelKey []byte

	AvailableKey    []byte
	SeqSendKey      []byte
	SeqRecvKey      []byte
	PacketCommitKey func(index uint64) []byte

	Connection connection.CLIObject

	Root merkle.Root
	Cdc  *codec.Codec
}

func (man Manager) CLIObject(root merkle.Root, connid, chanid string) CLIObject {
	obj := man.object(connid, chanid)
	return CLIObject{
		ChanID:     chanid,
		ChannelKey: obj.channel.Key(),

		AvailableKey: obj.available.Key(),
		SeqSendKey:   obj.seqsend.Key(),
		SeqRecvKey:   obj.seqrecv.Key(),
		PacketCommitKey: func(index uint64) []byte {
			return obj.packets.Value(index).Key()
		},

		Connection: man.connection.CLIObject(root, connid),

		Root: root,
		Cdc:  obj.channel.Cdc(),
	}
}

func (obj CLIObject) query(ctx context.CLIContext, key []byte, ptr interface{}) (merkle.Proof, error) {
	resp, err := ctx.QueryABCI(obj.Root.RequestQuery(key))
	if err != nil {
		return merkle.Proof{}, err
	}
	proof := merkle.Proof{
		Key:   key,
		Proof: resp.Proof,
	}
	err = obj.Cdc.UnmarshalBinaryBare(resp.Value, ptr)
	return proof, err

}

func (obj CLIObject) Channel(ctx context.CLIContext) (res Channel, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.ChannelKey, &res)
	return
}

func (obj CLIObject) SeqSend(ctx context.CLIContext) (res uint64, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.SeqSendKey, &res)
	return
}

func (obj CLIObject) SeqRecv(ctx context.CLIContext) (res uint64, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.SeqRecvKey, &res)
	return
}

func (obj CLIObject) PacketCommit(ctx context.CLIContext, index uint64) (res Packet, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.PacketCommitKey(index), &res)
	return
}

type CLIHandshakeObject struct {
	CLIObject

	StateKey   []byte
	TimeoutKey []byte
}

func (man Handshaker) CLIObject(root merkle.Root, connid, chanid string) CLIHandshakeObject {
	obj := man.object(man.man.object(connid, chanid))
	return CLIHandshakeObject{
		CLIObject: man.man.CLIObject(root, connid, chanid),

		StateKey:   obj.state.Key(),
		TimeoutKey: obj.nextTimeout.Key(),
	}
}

func (obj CLIHandshakeObject) State(ctx context.CLIContext) (res State, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.StateKey, &res)
	return
}

func (obj CLIHandshakeObject) Timeout(ctx context.CLIContext) (res uint64, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.TimeoutKey, &res)
	return
}
