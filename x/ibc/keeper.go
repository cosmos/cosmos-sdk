package ibc

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/mapping"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
)

type Keeper struct {
	client     client.Manager
	connection connection.Manager
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, cidgen client.IDGenerator) Keeper {
	base := mapping.NewBase(cdc, key)
	return newKeeper(base.Prefix([]byte{0x00}), base.Prefix([]byte{0x01}), cidgen)
}

func newKeeper(protocol mapping.Base, free mapping.Base, cidgen client.IDGenerator) (k Keeper) {
	k = Keeper{
		client: client.NewManager(
			protocol.Prefix([]byte("clients")),
			free.Prefix([]byte("clients")),
			cidgen,
		),
		connection: connection.NewManager(
			protocol.Prefix([]byte("connections")),
			free.Prefix([]byte("connections")),
		),
	}

	return
}
