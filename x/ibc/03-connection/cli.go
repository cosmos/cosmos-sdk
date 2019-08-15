package connection

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/client/context"

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

func (obj Object) ConnectionCLI(ctx context.CLIContext) (res Connection, proof merkle.Proof, err error) {
	tmproof, err := obj.Connection.Query(ctx, &res)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.Connection)
	return
}

func (obj Object) AvailableCLI(ctx context.CLIContext) (res bool, proof merkle.Proof, err error) {
	res, tmproof, err := obj.Available.Query(ctx)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.Available)
	return
}

func (obj Object) KindCLI(ctx context.CLIContext) (res string, proof merkle.Proof, err error) {
	res, tmproof, err := obj.Kind.Query(ctx)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.Kind)
	return
}

func (man Handshaker) CLIObject(connid, clientid string) HandshakeObject {
	return man.Object(man.man.CLIObject(connid, clientid))
}

func (man Handshaker) CLIQuery(ctx context.CLIContext, connid string) (HandshakeObject, error) {
	obj := man.man.Object(connid)
	conn, _, err := obj.ConnectionCLI(ctx)
	if err != nil {
		return HandshakeObject{}, err
	}
	obj.Client = man.man.client.Object(conn.Client)
	return man.Object(obj), nil
}

func (obj HandshakeObject) StateCLI(ctx context.CLIContext) (res byte, proof merkle.Proof, err error){
	res, tmproof, err := obj.State.Query(ctx)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.State)
	return
}

func (obj HandshakeObject) CounterpartyClientCLI(ctx context.CLIContext) (res string, proof merkle.Proof, err error)  {
	res, tmproof, err := obj.CounterpartyClient.Query(ctx)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.CounterpartyClient)
	return 
}

func (obj HandshakeObject) NextTimeoutCLI(ctx context.CLIContext) (res uint64, proof merkle.Proof, err error){
	res, tmproof, err := obj.NextTimeout.Query(ctx)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.NextTimeout)
	return
}
