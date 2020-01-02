package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	port "github.com/cosmos/cosmos-sdk/x/ibc/05-port"
	transfer "github.com/cosmos/cosmos-sdk/x/ibc/20-transfer"
)

// Keeper defines each ICS keeper for IBC
type Keeper struct {
	ClientKeeper     client.Keeper
	ConnectionKeeper connection.Keeper
	ChannelKeeper    channel.Keeper
	PortKeeper       port.Keeper
	TransferKeeper   transfer.Keeper
}

// NewKeeper creates a new ibc Keeper
func NewKeeper(
	cdc *codec.Codec, key sdk.StoreKey,
	bk transfer.BankKeeper, sk transfer.SupplyKeeper,
) Keeper {
	clientKeeper := client.NewKeeper(cdc, key)
	connectionKeeper := connection.NewKeeper(cdc, key, clientKeeper)
	portKeeper := port.NewKeeper(cdc, key)
	channelKeeper := channel.NewKeeper(cdc, key, clientKeeper, connectionKeeper, portKeeper)

	// TODO: move out of IBC keeper. Blocked on ADR15
	capKey := portKeeper.BindPort(bank.ModuleName)
	transferKeeper := transfer.NewKeeper(
		cdc, key, capKey,
		clientKeeper, connectionKeeper, channelKeeper, bk, sk,
	)

	return Keeper{
		ClientKeeper:     clientKeeper,
		ConnectionKeeper: connectionKeeper,
		ChannelKeeper:    channelKeeper,
		PortKeeper:       portKeeper,
		TransferKeeper:   transferKeeper,
	}
}
