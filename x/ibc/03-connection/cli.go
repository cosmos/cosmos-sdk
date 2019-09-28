package connection

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/store/state"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

func (man Manager) CLIState(connid, clientid string) State {
	obj := man.State(connid)
	obj.Client = man.client.State(clientid)
	return obj
}

func (man Manager) CLIQuery(q state.ABCIQuerier, connid string) (State, error) {
	obj := man.State(connid)
	conn, _, err := obj.ConnectionCLI(q)
	if err != nil {
		return State{}, err
	}
	obj.Client = man.client.State(conn.Client)
	return obj, nil
}

func (obj State) prefix() []byte {
	return bytes.Split(obj.Connection.KeyBytes(), LocalRoot())[0]
}

func (obj State) ConnectionCLI(q state.ABCIQuerier) (res Connection, proof merkle.Proof, err error) {
	tmproof, err := obj.Connection.Query(q, &res)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.Connection)
	return
}

func (obj State) AvailableCLI(q state.ABCIQuerier) (res bool, proof merkle.Proof, err error) {
	res, tmproof, err := obj.Available.Query(q)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.Available)
	return
}

func (obj State) KindCLI(q state.ABCIQuerier) (res string, proof merkle.Proof, err error) {
	res, tmproof, err := obj.Kind.Query(q)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.Kind)
	return
}

func (man Handshaker) CLIState(connid, clientid string) HandshakeState {
	return man.CreateState(man.man.CLIState(connid, clientid))
}

func (man Handshaker) CLIQuery(q state.ABCIQuerier, connid string) (HandshakeState, error) {
	state, err := man.man.CLIQuery(q, connid)
	if err != nil {
		return HandshakeState{}, err
	}
	return man.CreateState(state), nil
}

func (obj HandshakeState) StageCLI(q state.ABCIQuerier) (res byte, proof merkle.Proof, err error) {
	res, tmproof, err := obj.Stage.Query(q)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.Stage)
	return
}

func (obj HandshakeState) CounterpartyClientCLI(q state.ABCIQuerier) (res string, proof merkle.Proof, err error) {
	res, tmproof, err := obj.CounterpartyClient.Query(q)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.CounterpartyClient)
	return
}
