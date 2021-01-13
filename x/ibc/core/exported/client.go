package exported

import (
	ics23 "github.com/confio/ics23/go"
	proto "github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// TypeClientMisbehaviour is the shared evidence misbehaviour type
	TypeClientMisbehaviour string = "client_misbehaviour"

	// Solomachine is used to indicate that the light client is a solo machine.
	Solomachine string = "06-solomachine"

	// Tendermint is used to indicate that the client uses the Tendermint Consensus Algorithm.
	Tendermint string = "07-tendermint"

	// Localhost is the client type for a localhost client. It is also used as the clientID
	// for the localhost client.
	Localhost string = "09-localhost"
)

// ClientState defines the required common functions for light clients.
type ClientState interface {
	proto.Message

	ClientType() string
	GetLatestHeight() Height
	IsFrozen() bool
	GetFrozenHeight() Height
	Validate() error
	GetProofSpecs() []*ics23.ProofSpec

	// Initialization function
	// Clients must validate the initial consensus state, and may store any client-specific metadata
	// necessary for correct light client operation
	Initialize(sdk.Context, codec.BinaryMarshaler, sdk.KVStore, ConsensusState) error

	// Genesis function
	ExportMetadata(sdk.KVStore) []GenesisMetadata

	// Update and Misbehaviour functions

	CheckHeaderAndUpdateState(sdk.Context, codec.BinaryMarshaler, sdk.KVStore, Header) (ClientState, ConsensusState, error)
	CheckMisbehaviourAndUpdateState(sdk.Context, codec.BinaryMarshaler, sdk.KVStore, Misbehaviour) (ClientState, error)
	CheckProposedHeaderAndUpdateState(sdk.Context, codec.BinaryMarshaler, sdk.KVStore, Header) (ClientState, ConsensusState, error)

	// Upgrade functions
	// NOTE: proof heights are not included as upgrade to a new revision is expected to pass only on the last
	// height committed by the current revision. Clients are responsible for ensuring that the planned last
	// height of the current revision is somehow encoded in the proof verification process.
	// This is to ensure that no premature upgrades occur, since upgrade plans committed to by the counterparty
	// may be cancelled or modified before the last planned height.
	VerifyUpgradeAndUpdateState(
		ctx sdk.Context,
		cdc codec.BinaryMarshaler,
		store sdk.KVStore,
		newClient ClientState,
		newConsState ConsensusState,
		proofUpgradeClient,
		proofUpgradeConsState []byte,
	) (ClientState, ConsensusState, error)
	// Utility function that zeroes out any client customizable fields in client state
	// Ledger enforced fields are maintained while all custom fields are zero values
	// Used to verify upgrades
	ZeroCustomFields() ClientState

	// State verification functions

	VerifyClientState(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		height Height,
		prefix Prefix,
		counterpartyClientIdentifier string,
		proof []byte,
		clientState ClientState,
	) error
	VerifyClientConsensusState(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
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
		currentTimestamp uint64,
		delayPeriod uint64,
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
		currentTimestamp uint64,
		delayPeriod uint64,
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
		currentTimestamp uint64,
		delayPeriod uint64,
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
		currentTimestamp uint64,
		delayPeriod uint64,
		prefix Prefix,
		proof []byte,
		portID,
		channelID string,
		nextSequenceRecv uint64,
	) error
}

// ConsensusState is the state of the consensus process
type ConsensusState interface {
	proto.Message

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
	proto.Message

	ClientType() string
	GetClientID() string
	ValidateBasic() error

	// Height at which the infraction occurred
	GetHeight() Height
}

// Header is the consensus state update information
type Header interface {
	proto.Message

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
	GetRevisionNumber() uint64
	GetRevisionHeight() uint64
	Increment() Height
	Decrement() (Height, bool)
	String() string
}

// GenesisMetadata is a wrapper interface over clienttypes.GenesisMetadata
// all clients must use the concrete implementation in types
type GenesisMetadata interface {
	// return store key that contains metadata without clientID-prefix
	GetKey() []byte
	// returns metadata value
	GetValue() []byte
}
