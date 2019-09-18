package channel

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/store/state"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

func (man Manager) CLIObject(portid, chanid string, connids []string) Object {
	obj := man.object(portid, chanid)
	for _, connid := range connids {
		obj.Connections = append(obj.Connections, man.connection.Object(connid))
	}
	return obj
}

func (man Manager) CLIQuery(q state.ABCIQuerier, portid, chanid string) (obj Object, err error) {
	obj = man.object(portid, chanid)
	channel, _, err := obj.ChannelCLI(q)
	if err != nil {
		return
	}
	for _, connid := range channel.ConnectionHops {
		obj.Connections = append(obj.Connections, man.connection.Object(connid))
	}
	return
}

func (obj Object) prefix() []byte {
	return bytes.Split(obj.Channel.KeyBytes(), LocalRoot())[0]
}

func (obj Object) ChannelCLI(q state.ABCIQuerier) (res Channel, proof merkle.Proof, err error) {
	tmproof, err := obj.Channel.Query(q, &res)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.Channel)
	return
}

func (obj Object) AvailableCLI(q state.ABCIQuerier) (res bool, proof merkle.Proof, err error) {
	res, tmproof, err := obj.Available.Query(q)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.Available)
	return
}

func (obj Object) SeqSendCLI(q state.ABCIQuerier) (res uint64, proof merkle.Proof, err error) {
	res, tmproof, err := obj.SeqSend.Query(q)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.SeqSend)
	return
}

func (obj Object) SeqRecvCLI(q state.ABCIQuerier) (res uint64, proof merkle.Proof, err error) {
	res, tmproof, err := obj.SeqRecv.Query(q)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.SeqRecv)
	return
}

func (obj Object) PacketCLI(q state.ABCIQuerier, index uint64) (res Packet, proof merkle.Proof, err error) {
	packet := obj.Packets.Value(index)
	tmproof, err := packet.Query(q, &res)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), packet)
	return
}

func (man Handshaker) CLIQuery(q state.ABCIQuerier, portid, chanid string) (HandshakeObject, error) {
	obj, err := man.Manager.CLIQuery(q, portid, chanid)
	if err != nil {
		return HandshakeObject{}, err
	}
	return man.object(obj), nil
}

func (man Handshaker) CLIObject(portid, chanid string, connids []string) HandshakeObject {
	return man.object(man.Manager.CLIObject(portid, chanid, connids))
}

func (obj HandshakeObject) StateCLI(q state.ABCIQuerier) (res State, proof merkle.Proof, err error) {
	res, tmproof, err := obj.State.Query(q)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.State)
	return
}

func (obj HandshakeObject) NextTimeoutCLI(q state.ABCIQuerier) (res uint64, proof merkle.Proof, err error) {
	res, tmproof, err := obj.NextTimeout.Query(q)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.NextTimeout)
	return
}
