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

	Path merkle.Path
	Cdc  *codec.Codec
}

func (man Manager) cliObject(path merkle.Path, chanid, connid string) CLIObject {
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

		Path: path,
		Cdc:  obj.channel.Cdc(),
	}
}

func (man Manager) CLIQuery(ctx context.CLIContext, path merkle.Path, chanid, connid string) CLIObject {
	obj := man.cliObject(path, chanid, connid)
	obj.Connection = man.connection.CLIQuery(ctx, path, connid)
	return obj
}

func (man Manager) CLIObject(path merkle.Path, chanid, connid, clientid string) CLIObject {
	obj := man.cliObject(path, chanid, connid)
	obj.Connection = man.connection.CLIObject(path, connid, clientid)
	return obj
}

func (obj CLIObject) query(ctx context.CLIContext, key []byte, ptr interface{}) (merkle.Proof, error) {
	resp, err := ctx.QueryABCI(obj.Path.RequestQuery(key))
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

func (obj CLIObject) Available(ctx context.CLIContext) (res bool, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.AvailableKey, &res)
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

func (obj CLIObject) Packet(ctx context.CLIContext, index uint64) (res Packet, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.PacketCommitKey(index), &res)
	return
}

type CLIHandshakeObject struct {
	CLIObject

	StateKey   []byte
	TimeoutKey []byte
}

func (man Handshaker) CLIQuery(ctx context.CLIContext, path merkle.Path, chanid, connid string) CLIHandshakeObject {
	obj := man.object(man.man.object(connid, chanid))
	return CLIHandshakeObject{
		CLIObject: man.man.CLIQuery(ctx, path, chanid, connid),

		StateKey:   obj.state.Key(),
		TimeoutKey: obj.nextTimeout.Key(),
	}
}

func (man Handshaker) CLIObject(path merkle.Path, chanid, connid, clientid string) CLIHandshakeObject {
	obj := man.object(man.man.object(connid, chanid))
	return CLIHandshakeObject{
		CLIObject: man.man.CLIObject(path, chanid, connid, clientid),

		StateKey:   obj.state.Key(),
		TimeoutKey: obj.nextTimeout.Key(),
	}
}

func (obj CLIHandshakeObject) State(ctx context.CLIContext) (res State, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.StateKey, &res)
	return
}

func (obj CLIHandshakeObject) NextTimeout(ctx context.CLIContext) (res uint64, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.TimeoutKey, &res)
	return
}
