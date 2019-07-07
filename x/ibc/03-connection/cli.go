package connection

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

// CLIObject stores the key for each object fields
type CLIObject struct {
	ID            string
	ConnectionKey []byte
	SendableKey   []byte
	ReceivableKey []byte
	KindKey       []byte

	Client client.CLIObject

	Root merkle.Root
	Cdc  *codec.Codec
}

func (man Manager) CLIObject(root merkle.Root, id string) CLIObject {
	obj := man.object(id)
	return CLIObject{
		ID:            obj.id,
		ConnectionKey: obj.connection.Key(),
		SendableKey:   obj.sendable.Key(),
		ReceivableKey: obj.receivable.Key(),
		KindKey:       obj.kind.Key(),

		// TODO: unify man.CLIObject() <=> obj.CLI()
		Client: obj.client.CLI(root),

		Root: root,
		Cdc:  obj.connection.Cdc(),
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

func (obj CLIObject) Connection(ctx context.CLIContext) (res Connection, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.ConnectionKey, &res)
	return
}

func (obj CLIObject) Sendable(ctx context.CLIContext) (res bool, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.SendableKey, &res)
	return
}

func (obj CLIObject) Receivable(ctx context.CLIContext) (res bool, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.ReceivableKey, &res)
	return
}

func (obj CLIObject) Kind(ctx context.CLIContext) (res string, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.KindKey, &res)
	return
}

type CLIHandshakeObject struct {
	CLIObject

	StateKey              []byte
	CounterpartyClientKey []byte
	TimeoutKey            []byte
}

func (man Handshaker) CLIObject(root merkle.Root, id string) CLIHandshakeObject {
	obj := man.object(man.man.object(id))
	return CLIHandshakeObject{
		CLIObject: man.man.CLIObject(root, id),

		StateKey:              obj.state.Key(),
		CounterpartyClientKey: obj.counterpartyClient.Key(),
		TimeoutKey:            obj.nextTimeout.Key(),
	}
}

func (obj CLIHandshakeObject) State(ctx context.CLIContext, root merkle.Root) (res byte, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.StateKey, &res)
	return
}

func (obj CLIHandshakeObject) CounterpartyClient(ctx context.CLIContext, root merkle.Root) (res string, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.CounterpartyClientKey, &res)
	return
}

func (obj CLIHandshakeObject) Timeout(ctx context.CLIContext, root merkle.Root) (res uint64, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.TimeoutKey, &res)
	return
}

/*
func (obj CLIObject) State(ctx context.CLIContext, root merkle.Root) (res bool, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.StateKey, &res)
	return
}

func (obj CLIObject) NextTimeout(ctx context.CLIContext, root merkle.Root) (res time.Time, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.NextTimeoutKey, &res)
	return
}

func (obj CLIObject) Permission(ctx context.CLIContext, root merkle.Root) (res string, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.PermissionKey, &res)
	return
}
*/
