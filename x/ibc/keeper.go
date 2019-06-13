package ibc

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type Keeper struct {
	client     client.Manager
	connection connection.Manager

	cdc *codec.Codec
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, cidgen client.IDGenerator) Keeper {
	base := mapping.NewBase(cdc, key)
	return newKeeper(base.Prefix([]byte{0x00}), base.Prefix([]byte{0x01}), cidgen, cdc)
}

type ContextKeyRemoteKVStore struct{}

func newRemoteKeeper(cdc *codec.Codec) Keeper {
	return newKeeper(
		mapping.NewBaseWithGetter(cdc, commitment.GetStore),
		mapping.EmptyBase(), // Will cause panic when trying to access on non-protocol states
		nil,                 // cidgen should not be used as it is called when a new client is set
		cdc,
	)
}

func newKeeper(protocol mapping.Base, free mapping.Base, cidgen client.IDGenerator, cdc *codec.Codec) (k Keeper) {
	k = Keeper{
		client:     client.NewManager(protocol, free, cidgen),
		connection: connection.NewManager(protocol, free, k.client),
		cdc:        cdc,
	}

	return
}

func (k Keeper) Exec(ctx sdk.Context,
	connid string, proofs []commitment.Proof, fullProofs []commitment.FullProof,
	fn func(sdk.Context, Keeper) error,
) error {
	root, err := k.root(ctx, connid)
	if err != nil {
		return err
	}

	store, err := commitment.NewStore(root, proofs, fullProofs)
	if err != nil {
		return err
	}

	return fn(commitment.WithStore(ctx, store), newRemoteKeeper(k.cdc))
}

func (k Keeper) root(ctx sdk.Context, connid string) (commitment.Root, error) {
	conn, err := k.connection.Query(ctx, connid)
	if err != nil {
		return nil, err
	}
	client, err := k.client.Query(ctx, conn.ClientID())
	if err != nil {
		return nil, err
	}
	return client.Value(ctx).GetRoot(), nil
}
