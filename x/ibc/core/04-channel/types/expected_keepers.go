package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

// ClientKeeper expected account IBC client keeper
type ClientKeeper interface {
	GetClientState(ctx sdk.Context, clientID string) (exported.ClientState, bool)
	GetClientConsensusState(ctx sdk.Context, clientID string, height exported.Height) (exported.ConsensusState, bool)
}

// ConnectionKeeper expected account IBC connection keeper
type ConnectionKeeper interface {
	GetConnection(ctx sdk.Context, connectionID string) (connectiontypes.ConnectionEnd, bool)
	GetTimestampAtHeight(
		ctx sdk.Context,
		connection connectiontypes.ConnectionEnd,
		height exported.Height,
	) (uint64, error)
	VerifyChannelState(
		ctx sdk.Context,
		connection exported.ConnectionI,
		height exported.Height,
		proof []byte,
		portID,
		channelID string,
		channel exported.ChannelI,
	) error
	VerifyPacketCommitment(
		ctx sdk.Context,
		connection exported.ConnectionI,
		height exported.Height,
		proof []byte,
		portID,
		channelID string,
		sequence uint64,
		commitmentBytes []byte,
	) error
	VerifyPacketAcknowledgement(
		ctx sdk.Context,
		connection exported.ConnectionI,
		height exported.Height,
		proof []byte,
		portID,
		channelID string,
		sequence uint64,
		acknowledgement []byte,
	) error
	VerifyPacketReceiptAbsence(
		ctx sdk.Context,
		connection exported.ConnectionI,
		height exported.Height,
		proof []byte,
		portID,
		channelID string,
		sequence uint64,
	) error
	VerifyNextSequenceRecv(
		ctx sdk.Context,
		connection exported.ConnectionI,
		height exported.Height,
		proof []byte,
		portID,
		channelID string,
		nextSequenceRecv uint64,
	) error
}

// PortKeeper expected account IBC port keeper
type PortKeeper interface {
	Authenticate(ctx sdk.Context, key *capabilitytypes.Capability, portID string) bool
}
