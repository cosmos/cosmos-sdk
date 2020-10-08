package exported

import (
	ics23 "github.com/confio/ics23/go"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

const (
	// TypeClientMisbehaviour is the shared evidence misbehaviour type
	TypeClientMisbehaviour string = "client_misbehaviour"

	// Localhost is the client type for a localhost client. It is also used as the clientID
	// for the localhost client.
	Localhost string = "localhost"
)

// ClientState defines the required common functions for light clients.
type ClientState interface {
	ClientType() string
	GetLatestHeight() Height
	IsFrozen() bool
	GetFrozenHeight() Height
	Validate() error
	GetProofSpecs() []*ics23.ProofSpec

	// Update and Misbehaviour functions

	CheckHeaderAndUpdateState(sdk.Context, codec.BinaryMarshaler, sdk.KVStore, Header) (ClientState, ConsensusState, error)
	CheckMisbehaviourAndUpdateState(sdk.Context, codec.BinaryMarshaler, sdk.KVStore, Misbehaviour) (ClientState, error)
	CheckProposedHeaderAndUpdateState(sdk.Context, codec.BinaryMarshaler, sdk.KVStore, Header) (ClientState, ConsensusState, error)

	// Upgrade functions
	VerifyUpgrade(
		ctx sdk.Context,
		cdc codec.BinaryMarshaler,
		store sdk.KVStore,
		newClient ClientState,
		upgradeHeight Height,
		proofUpgrade []byte,
	) error
	// Utility function that zeroes out any client customizable fields in client state
	// Ledger enforced fields are maintained while all custom fields are zero values
	// Used to verify upgrades
	ZeroCustomFields() ClientState

	// State verification functions

	VerifyClientState(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		root Root,
		height Height,
		prefix Prefix,
		counterpartyClientIdentifier string,
		proof []byte,
		clientState ClientState,
	) error
	VerifyClientConsensusState(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		root Root,
		height Height,
		counterpartyClientIdentifier string,
		consensusHeight Height,
		prefix Prefix,
		proof []byte,
		consensusState ConsensusState,
	) error
	VerifyConnectionState(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		height Height,
		prefix Prefix,
		proof []byte,
		connectionID string,
		connectionEnd ConnectionI,
	) error
	VerifyChannelState(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		height Height,
		prefix Prefix,
		proof []byte,
		portID,
		channelID string,
		channel ChannelI,
	) error
	VerifyPacketCommitment(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		height Height,
		prefix Prefix,
		proof []byte,
		portID,
		channelID string,
		sequence uint64,
		commitmentBytes []byte,
	) error
	VerifyPacketAcknowledgement(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		height Height,
		prefix Prefix,
		proof []byte,
		portID,
		channelID string,
		sequence uint64,
		acknowledgement []byte,
	) error
	VerifyPacketReceiptAbsence(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		height Height,
		prefix Prefix,
		proof []byte,
		portID,
		channelID string,
		sequence uint64,
	) error
	VerifyNextSequenceRecv(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		height Height,
		prefix Prefix,
		proof []byte,
		portID,
		channelID string,
		nextSequenceRecv uint64,
	) error
}

// ConsensusState is the state of the consensus process
type ConsensusState interface {
	ClientType() string // Consensus kind

	// GetRoot returns the commitment root of the consensus state,
	// which is used for key-value pair verification.
	GetRoot() Root

	// GetTimestamp returns the timestamp (in nanoseconds) of the consensus state
	GetTimestamp() uint64

	ValidateBasic() error
}

// Misbehaviour defines counterparty misbehaviour for a specific consensus type
type Misbehaviour interface {
	ClientType() string
	GetClientID() string
	String() string
	ValidateBasic() error

	// Height at which the infraction occurred
	GetHeight() Height
}

// Header is the consensus state update information
type Header interface {
	ClientType() string
	GetHeight() Height
	ValidateBasic() error
}

// Height is a wrapper interface over clienttypes.Height
// all clients must use the concrete implementation in types
type Height interface {
	IsZero() bool
	LT(Height) bool
	LTE(Height) bool
	EQ(Height) bool
	GT(Height) bool
	GTE(Height) bool
	GetVersionNumber() uint64
	GetVersionHeight() uint64
	Decrement() (Height, bool)
	String() string
}
