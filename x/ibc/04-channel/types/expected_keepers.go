package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// ClientKeeper expected account IBC client keeper
type ClientKeeper interface {
	GetClientState(ctx sdk.Context, clientID string) (clientexported.ClientState, bool)
	GetConsensusState(ctx sdk.Context, clientID string) (clientexported.ConsensusState, bool)
}

// ConnectionKeeper expected account IBC connection keeper
type ConnectionKeeper interface {
	GetConnection(ctx sdk.Context, connectionID string) (connectionexported.ConnectionI, bool)
	VerifyChannelState(
		ctx sdk.Context,
		connection connectionexported.ConnectionI,
		height uint64,
		proof commitment.ProofI,
		portID,
		channelID string,
		channel Channel,
		consensusState clientexported.ConsensusState,
	) error
	VerifyPacketCommitment(
		ctx sdk.Context,
		connection connectionexported.ConnectionI,
		height uint64,
		proof commitment.ProofI,
		portID,
		channelID string,
		sequence uint64,
		commitmentBytes []byte,
		consensusState clientexported.ConsensusState,
	) error
	VerifyPacketAcknowledgement(
		ctx sdk.Context,
		connection connectionexported.ConnectionI,
		height uint64,
		proof commitment.ProofI,
		portID,
		channelID string,
		sequence uint64,
		acknowledgement []byte,
		consensusState clientexported.ConsensusState,
	) error
	VerifyPacketAcknowledgementAbsence(
		ctx sdk.Context,
		connection connectionexported.ConnectionI,
		height uint64,
		proof commitment.ProofI,
		portID,
		channelID string,
		sequence uint64,
		consensusState clientexported.ConsensusState,
	) error
	VerifyNextSequenceRecv(
		ctx sdk.Context,
		connection connectionexported.ConnectionI,
		height uint64,
		proof commitment.ProofI,
		portID,
		channelID string,
		nextSequenceRecv uint64,
		consensusState clientexported.ConsensusState,
	) error
}

// PortKeeper expected account IBC port keeper
type PortKeeper interface {
	Authenticate(key sdk.CapabilityKey, portID string) bool
}
