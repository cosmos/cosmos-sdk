package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

// ClientKeeper expected account IBC client keeper
type ClientKeeper interface {
	GetClientState(ctx sdk.Context, clientID string) (clientexported.ClientState, bool)
}

// ConnectionKeeper expected account IBC connection keeper
type ConnectionKeeper interface {
	GetConnection(ctx sdk.Context, connectionID string) (connectiontypes.ConnectionEnd, bool)
	GetTimestampAtHeight(
		ctx sdk.Context,
		connection connectiontypes.ConnectionEnd,
		height uint64,
	) (uint64, error)
	VerifyChannelState(
		ctx sdk.Context,
		connection connectionexported.ConnectionI,
		height uint64,
		proof commitmentexported.Proof,
		portID,
		channelID string,
		channel exported.ChannelI,
	) error
	VerifyPacketCommitment(
		ctx sdk.Context,
		connection connectionexported.ConnectionI,
		height uint64,
		proof commitmentexported.Proof,
		portID,
		channelID string,
		sequence uint64,
		commitmentBytes []byte,
	) error
	VerifyPacketAcknowledgement(
		ctx sdk.Context,
		connection connectionexported.ConnectionI,
		height uint64,
		proof commitmentexported.Proof,
		portID,
		channelID string,
		sequence uint64,
		acknowledgement []byte,
	) error
	VerifyPacketAcknowledgementAbsence(
		ctx sdk.Context,
		connection connectionexported.ConnectionI,
		height uint64,
		proof commitmentexported.Proof,
		portID,
		channelID string,
		sequence uint64,
	) error
	VerifyNextSequenceRecv(
		ctx sdk.Context,
		connection connectionexported.ConnectionI,
		height uint64,
		proof commitmentexported.Proof,
		portID,
		channelID string,
		nextSequenceRecv uint64,
	) error
}

// PortKeeper expected account IBC port keeper
type PortKeeper interface {
	Authenticate(ctx sdk.Context, key *capability.Capability, portID string) bool
}
