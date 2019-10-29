package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"
)

// BankKeeper expected bank keeper
type BankKeeper interface {
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) sdk.Error
}

// ChannelKeeper expected IBC channel keeper
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel channel.Channel, found bool)
	GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool)
	SendPacket(ctx sdk.Context, packet channelexported.PacketI, portCapability sdk.CapabilityKey) sdk.Error
}

// SupplyKeeper expected supply keeper
type SupplyKeeper interface {
	GetModuleAccount(ctx sdk.Context, name string) supplyexported.ModuleAccountI

	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) sdk.Error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) sdk.Error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string,
		recipientAddr sdk.AccAddress, amt sdk.Coins) sdk.Error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress,
		recipientModule string, amt sdk.Coins) sdk.Error
}
