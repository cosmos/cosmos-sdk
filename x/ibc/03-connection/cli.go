package connection

import (
	"time"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

// CLIObject stores the key for each object fields
type CLIObject struct {
	ID             string
	ConnectionKey  []byte
	StateKey       []byte
	NextTimeoutKey []byte

	Client client.CLIObject

	Root merkle.Root
	Cdc  *codec.Codec
}

func (man Manager) CLIObject(root merkle.Root, id string) CLIObject {
	obj := man.object(id)
	return CLIObject{
		ID:             obj.id,
		ConnectionKey:  obj.connection.Key(),
		StateKey:       obj.state.Key(),
		NextTimeoutKey: obj.nextTimeout.Key(),

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

func (obj CLIObject) Connection(ctx context.CLIContext, root merkle.Root) (res Connection, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.ConnectionKey, &res)
	return
}

func (obj CLIObject) State(ctx context.CLIContext, root merkle.Root) (res bool, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.StateKey, &res)
	return
}

func (obj CLIObject) NextTimeout(ctx context.CLIContext, root merkle.Root) (res time.Time, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.NextTimeoutKey, &res)
	return
}
