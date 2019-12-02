package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	// connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// ConnectionKeeper expected account IBC connection keeper
type ConnectionKeeper interface {
	// GetConnection(ctx sdk.Context, connectionID string) (connectiontypes.ConnectionEnd, bool)
	VerifyChannelState(
		ctx sdk.Context,
		clientID string, // i.e connection.ClientID
		height uint64,
		prefix commitment.PrefixI,
		proof commitment.ProofI,
		portID,
		channelID string,
		channel Channel,
	) (bool, error)
	VerifyPacketCommitment(
		ctx sdk.Context,
		clientID string, // i.e connection.ClientID
		height uint64,
		prefix commitment.PrefixI,
		proof commitment.ProofI,
		portID,
		channelID string,
		sequence uint64,
		commitmentBz []byte,
	) (bool, error)
	VerifyPacketAcknowledgement(
		ctx sdk.Context,
		clientID string, // i.e connection.ClientID
		height uint64,
		prefix commitment.PrefixI,
		proof commitment.ProofI,
		portID,
		channelID string,
		sequence uint64,
		acknowledgement []byte,
	) (bool, error)
	VerifyPacketAcknowledgementAbsence(
		ctx sdk.Context,
		clientID string, // i.e connection.ClientID
		height uint64,
		prefix commitment.PrefixI,
		proof commitment.ProofI,
		portID,
		channelID string,
		sequence uint64,
	) (bool, error)
	VerifyNextSequenceRecv(
		ctx sdk.Context,
		clientID string, // i.e connection.ClientID
		height uint64,
		prefix commitment.PrefixI,
		proof commitment.ProofI,
		portID,
		channelID string,
		nextSequenceRecv uint64,
	) (bool, error)
}

// PortKeeper expected account IBC port keeper
type PortKeeper interface {
	Authenticate(key sdk.CapabilityKey, portID string) bool
}
