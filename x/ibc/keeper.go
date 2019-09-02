package ibc

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	"github.com/cosmos/cosmos-sdk/x/ibc/version"
)

type Keeper struct {
	client     client.Manager
	connection connection.Handshaker
	channel    channel.Handshaker
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey) Keeper {
	base := state.NewMapping(key, cdc, version.DefaultPrefix())
	climan := client.NewManager(base)
	connman := connection.NewManager(base, climan)
	chanman := channel.NewManager(base, connman)
	return Keeper{
		client:     climan,
		connection: connection.NewHandshaker(connman),
		channel:    channel.NewHandshaker(chanman),
	}
}
