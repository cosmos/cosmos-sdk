package ibc

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
)

type Keeper struct {
	client     client.Manager
	connection connection.Handshaker
	channel    channel.Handshaker
}

func (k Keeper) Client() client.Manager {
	return k.client
}

func (k Keeper) Connection() connection.Manager {
	return k.connection.Manager()
}

func (k Keeper) Channel() channel.Manager {
	return k.channel.Manager()
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey) Keeper {
	base := state.NewBase(cdc, key, []byte("v1"))
	climan := client.NewManager(base)
	connman := connection.NewManager(base, climan)
	chanman := channel.NewManager(base, connman)
	return Keeper{
		client:     climan,
		connection: connection.NewHandshaker(connman),
		channel:    channel.NewHandshaker(chanman),
	}
}
