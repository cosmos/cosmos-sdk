package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/capability"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

type IBCModule interface {
	OnChanOpenInit(
		ctx sdk.Context,
		order channelexported.Order,
		connectionHops []string,
		portID string,
		channelID string,
		counterParty channeltypes.Counterparty,
		version string,
	) error

	OnChanOpenTry(
		ctx sdk.Context,
		order channelexported.Order,
		connectionHops []string,
		portID,
		channelID string,
		counterparty channeltypes.Counterparty,
		version,
		counterpartyVersion string,
	) error

	OnChanOpenAck(
		ctx sdk.Context,
		portID,
		channelID string,
		counterpartyVersion string,
	) error

	OnChanOpenConfirm(
		ctx sdk.Context,
		portID,
		channelID string,
	) error

	OnChanCloseInit(
		ctx sdk.Context,
		portID,
		channelID string,
	) error

	OnChanCloseConfirm(
		ctx sdk.Context,
		portID,
		channelID string,
	) error

	OnRecvPacket(
		packet channeltypes.Packet,
	) error

	OnAcknowledgementPacket(
		packet channeltypes.Packet,
		acknowledment []byte,
	) error

	OnTimeoutPacket(
		packet channeltypes.Packet,
	) error

	GetCapability(ctx sdk.Context, name string) (capability.Capability, bool)

	ClaimCapability(ctx sdk.Context, cap capability.Capability, name string) error
}
