package channel

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/store/state"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

func (man Manager) CLIState(portid, chanid string, connids []string) State {
	obj := man.object(portid, chanid)
	for _, connid := range connids {
		obj.Connections = append(obj.Connections, man.connection.State(connid))
	}
	return obj
}

func (man Manager) CLIQuery(q state.ABCIQuerier, portid, chanid string) (obj State, err error) {
	obj = man.object(portid, chanid)
	channel, _, err := obj.ChannelCLI(q)
	if err != nil {
		return
	}
	for _, connid := range channel.ConnectionHops {
		obj.Connections = append(obj.Connections, man.connection.State(connid))
	}
	return
}

func (obj State) prefix() []byte {
	return bytes.Split(obj.Channel.KeyBytes(), LocalRoot())[0]
}

func (obj State) ChannelCLI(q state.ABCIQuerier) (res Channel, proof merkle.Proof, err error) {
	tmproof, err := obj.Channel.Query(q, &res)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.Channel)
	return
}

func (obj State) AvailableCLI(q state.ABCIQuerier) (res bool, proof merkle.Proof, err error) {
	res, tmproof, err := obj.Available.Query(q)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.Available)
	return
}

func (obj State) SeqSendCLI(q state.ABCIQuerier) (res uint64, proof merkle.Proof, err error) {
	res, tmproof, err := obj.SeqSend.Query(q)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.SeqSend)
	return
}

func (obj State) SeqRecvCLI(q state.ABCIQuerier) (res uint64, proof merkle.Proof, err error) {
	res, tmproof, err := obj.SeqRecv.Query(q)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.SeqRecv)
	return
}

func (obj State) PacketCLI(q state.ABCIQuerier, index uint64) (res Packet, proof merkle.Proof, err error) {
	packet := obj.Packets.Value(index)
	tmproof, err := packet.Query(q, &res)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), packet)
	return
}

func (man Handshaker) CLIQuery(q state.ABCIQuerier, portid, chanid string) (HandshakeState, error) {
	obj, err := man.Manager.CLIQuery(q, portid, chanid)
	if err != nil {
		return HandshakeState{}, err
	}
	return man.createState(obj), nil
}

func (man Handshaker) CLIState(portid, chanid string, connids []string) HandshakeState {
	return man.createState(man.Manager.CLIState(portid, chanid, connids))
}

func (obj HandshakeState) StageCLI(q state.ABCIQuerier) (res Stage, proof merkle.Proof, err error) {
	res, tmproof, err := obj.Stage.Query(q)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.Stage)
	return
}
