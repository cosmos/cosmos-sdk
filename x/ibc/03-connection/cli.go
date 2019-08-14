package connection

import (
	"github.com/cosmos/cosmos-sdk/client/context"

	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

func (man Manager) CLIObject(connid, clientid string) Object {
	obj := man.Object(connid)
	obj.Client = man.client.Object(clientid)
	return obj
}

func (obj Object) ConnectionCLI(ctx context.CLIContext, path merkle.Path) (res Connection, proof merkle.Proof, err error) {
	tmproof, err := obj.Connection.Query(ctx, &res)
	proof = merkle.NewProofFromValue(tmproof, path, obj.Connection)
	return
}

func (obj Object) AvailableCLI(ctx context.CLIContext, path merkle.Path) (res bool, proof merkle.Proof, err error) {
	res, tmproof, err := obj.Available.Query(ctx)
	proof = merkle.NewProofFromValue(tmproof, path, obj.Available)
	return
}

func (obj Object) KindCLI(ctx context.CLIContext, path merkle.Path) (res string, proof merkle.Proof, err error) {
	res, tmproof, err := obj.Kind.Query(ctx)
	proof = merkle.NewProofFromValue(tmproof, path, obj.Kind)
	return
}

func (man Handshaker) CLIObject(connid, clientid string) HandshakeObject {
	return man.Object(man.man.CLIObject(connid, clientid))
}

func (obj HandshakeObject) StateCLI(ctx context.CLIContext, path merkle.Path) (res byte, proof merkle.Proof, err error){
	res, tmproof, err := obj.State.Query(ctx)
	proof = merkle.NewProofFromValue(tmproof, path, obj.State)
	return
}

func (obj HandshakeObject) CounterpartyClientCLI(ctx context.CLIContext, path merkle.Path) (res string, proof merkle.Proof, err error)  {
	res, tmproof, err := obj.CounterpartyClient.Query(ctx)
	proof = merkle.NewProofFromValue(tmproof, path, obj.CounterpartyClient)
	return 
}

func (obj HandshakeObject) NextTimeoutCLI(ctx context.CLIContext, path merkle.Path) (res uint64, proof merkle.Proof,  err error){
	res, tmproof, err := obj.NextTimeout.Query(ctx)
	proof = merkle.NewProofFromValue(tmproof, path, obj.NextTimeout)
	return
}
