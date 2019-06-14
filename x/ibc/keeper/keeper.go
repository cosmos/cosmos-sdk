package ibc

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type Keeper struct {
	Client     client.Manager
	Connection connection.Manager
	Channel    channel.Manager

	cdc *codec.Codec
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, cidgen client.IDGenerator) Keeper {
	base := state.NewBase(cdc, key)
	return newKeeper(base.Prefix([]byte{0x00}), base.Prefix([]byte{0x01}), cidgen, cdc)
}

func DummyKeeper() Keeper {
	base := state.EmptyBase()
	return newKeeper(base.Prefix([]byte{0x00}), base.Prefix([]byte{0x01}), nil, nil)
}

type ContextKeyRemoteKVStore struct{}

func newKeeper(protocol state.Base, free state.Base, cidgen client.IDGenerator, cdc *codec.Codec) (k Keeper) {
	k = Keeper{
		Client:     client.NewManager(protocol, free, cidgen),
		Connection: connection.NewManager(protocol, free, k.Client),
		Channel:    channel.NewManager(protocol, k.Connection),
		cdc:        cdc,
	}

	return
}

func (k Keeper) ProofExec(ctx sdk.Context,
	connid string, proofs []commitment.Proof,
	fn func(sdk.Context) error,
) error {
	root, err := k.root(ctx, connid)
	if err != nil {
		return err
	}

	store, err := commitment.NewStore(root, proofs)
	if err != nil {
		return err
	}

	return fn(commitment.WithStore(ctx, store))
}

func (k Keeper) root(ctx sdk.Context, connid string) (commitment.Root, error) {
	conn, err := k.Connection.Query(ctx, connid)
	if err != nil {
		return nil, err
	}
	client, err := k.Client.Query(ctx, conn.ClientID())
	if err != nil {
		return nil, err
	}
	return client.Value(ctx).GetRoot(), nil
}
