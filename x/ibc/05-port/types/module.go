package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/capability"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// IBCModule defines an interface that implements all the callbacks
// that modules must define as specified in ICS-26
type IBCModule interface {
	OnChanOpenInit(
		ctx sdk.Context,
		order channeltypes.Order,
		connectionHops []string,
		portID string,
		channelID string,
		channelCap *capability.Capability,
		counterParty channeltypes.Counterparty,
		version string,
	) error

	OnChanOpenTry(
		ctx sdk.Context,
		order channeltypes.Order,
		connectionHops []string,
		portID,
		channelID string,
		channelCap *capability.Capability,
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
		ctx sdk.Context,
		packet channeltypes.Packet,
	) (*sdk.Result, error)

	OnAcknowledgementPacket(
		ctx sdk.Context,
		packet channeltypes.Packet,
		acknowledgement []byte,
	) (*sdk.Result, error)

	OnTimeoutPacket(
		ctx sdk.Context,
		packet channeltypes.Packet,
	) (*sdk.Result, error)
}
