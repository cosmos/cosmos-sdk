package connection

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/store/state"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

func (man Manager) CLIObject(connid, clientid string) Object {
	obj := man.Object(connid)
	obj.Client = man.client.Object(clientid)
	return obj
}

func (obj Object) prefix() []byte {
	return bytes.Split(obj.Connection.KeyBytes(), LocalRoot())[0]
}

func (obj Object) ConnectionCLI(q state.ABCIQuerier) (res Connection, proof merkle.Proof, err error) {
	tmproof, err := obj.Connection.Query(q, &res)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.Connection)
	return
}

func (obj Object) AvailableCLI(q state.ABCIQuerier) (res bool, proof merkle.Proof, err error) {
	res, tmproof, err := obj.Available.Query(q)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.Available)
	return
}

func (obj Object) KindCLI(q state.ABCIQuerier) (res string, proof merkle.Proof, err error) {
	res, tmproof, err := obj.Kind.Query(q)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.Kind)
	return
}

func (man Handshaker) CLIObject(connid, clientid string) HandshakeObject {
	return man.Object(man.man.CLIObject(connid, clientid))
}

func (man Handshaker) CLIQuery(q state.ABCIQuerier, connid string) (HandshakeObject, error) {
	obj := man.man.Object(connid)
	conn, _, err := obj.ConnectionCLI(q)
	if err != nil {
		return HandshakeObject{}, err
	}
	obj.Client = man.man.client.Object(conn.Client)
	return man.Object(obj), nil
}

func (obj HandshakeObject) StateCLI(q state.ABCIQuerier) (res byte, proof merkle.Proof, err error) {
	res, tmproof, err := obj.State.Query(q)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.State)
	return
}

func (obj HandshakeObject) CounterpartyClientCLI(q state.ABCIQuerier) (res string, proof merkle.Proof, err error) {
	res, tmproof, err := obj.CounterpartyClient.Query(q)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.CounterpartyClient)
	return
}
