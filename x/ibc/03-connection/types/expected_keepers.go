package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// ClientKeeper expected account IBC client keeper
type ClientKeeper interface {
	// GetConsensusState(ctx sdk.Context, clientID string) (clientexported.ConsensusState, bool)
	GetClientState(ctx sdk.Context, clientID string) (clienttypes.State, bool)
	VerifyClientConsensusState(
		ctx sdk.Context,
		clientState clienttypes.State,
		height uint64,
		prefix commitment.PrefixI,
		proof commitment.ProofI,
		consensusState clientexported.ConsensusState,
	) (bool, error)
	VerifyConnectionState(
		ctx sdk.Context,
		clientState clienttypes.State,
		height uint64,
		prefix commitment.PrefixI,
		proof commitment.ProofI,
		connectionID string,
		connection ConnectionEnd,
	) (bool, error)
	VerifyChannelState(
		ctx sdk.Context,
		clientState clienttypes.State,
		height uint64,
		prefix commitment.PrefixI,
		proof commitment.ProofI,
		portID,
		channelID string,
		channel channeltypes.Channel,
	) (bool, error)
	VerifyPacketCommitment(
		ctx sdk.Context,
		clientState clienttypes.State,
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
		clientState clienttypes.State,
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
		clientState clienttypes.State,
		height uint64,
		prefix commitment.PrefixI,
		proof commitment.ProofI,
		portID,
		channelID string,
		sequence uint64,
	) (bool, error)
	VerifyNextSequenceRecv(
		ctx sdk.Context,
		clientState clienttypes.State,
		height uint64,
		prefix commitment.PrefixI,
		proof commitment.ProofI,
		portID,
		channelID string,
		nextSequenceRecv uint64,
	) (bool, error)
}
