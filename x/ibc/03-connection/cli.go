package ics03

// import (
// 	"bytes"

// 	"github.com/cosmos/cosmos-sdk/store/state"
// 	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
// )

// func (man Manager) CLIState(connid, clientid string) State {
// 	state := man.State(connid)
// 	state.Client = man.client.State(clientid)
// 	return state
// }

// func (man Manager) CLIQuery(q state.ABCIQuerier, connid string) (State, error) {
// 	state := man.State(connid)
// 	conn, _, err := state.ConnectionCLI(q)
// 	if err != nil {
// 		return State{}, err
// 	}
// 	state.Client = man.client.State(conn.Client)
// 	return state, nil
// }

// func (state State) prefix() []byte {
// 	return bytes.Split(state.Connection.KeyBytes(), LocalRoot())[0]
// }

// func (state State) ConnectionCLI(q state.ABCIQuerier) (res Connection, proof merkle.Proof, err error) {
// 	tmproof, err := state.Connection.Query(q, &res)
// 	proof = merkle.NewProofFromValue(tmproof, state.prefix(), state.Connection)
// 	return
// }

// func (state State) AvailableCLI(q state.ABCIQuerier) (res bool, proof merkle.Proof, err error) {
// 	res, tmproof, err := state.Available.Query(q)
// 	proof = merkle.NewProofFromValue(tmproof, state.prefix(), state.Available)
// 	return
// }

// func (state State) KindCLI(q state.ABCIQuerier) (res string, proof merkle.Proof, err error) {
// 	res, tmproof, err := state.Kind.Query(q)
// 	proof = merkle.NewProofFromValue(tmproof, state.prefix(), state.Kind)
// 	return
// }

// func (man Handshaker) CLIState(connid, clientid string) HandshakeState {
// 	return man.CreateState(man.man.CLIState(connid, clientid))
// }

// func (man Handshaker) CLIQuery(q state.ABCIQuerier, connid string) (HandshakeState, error) {
// 	state, err := man.man.CLIQuery(q, connid)
// 	if err != nil {
// 		return HandshakeState{}, err
// 	}
// 	return man.CreateState(state), nil
// }

// func (state HandshakeState) StageCLI(q state.ABCIQuerier) (res byte, proof merkle.Proof, err error) {
// 	res, tmproof, err := state.Stage.Query(q)
// 	proof = merkle.NewProofFromValue(tmproof, state.prefix(), state.Stage)
// 	return
// }

// func (state HandshakeState) CounterpartyClientCLI(q state.ABCIQuerier) (res string, proof merkle.Proof, err error) {
// 	res, tmproof, err := state.CounterpartyClient.Query(q)
// 	proof = merkle.NewProofFromValue(tmproof, state.prefix(), state.CounterpartyClient)
// 	return
// }
