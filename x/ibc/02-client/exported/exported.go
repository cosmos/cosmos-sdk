package exported

import (
	"encoding/json"

	ics23 "github.com/confio/ics23/go"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

// ClientState defines the required common functions for light clients.
type ClientState interface {
	GetChainID() string
	ClientType() ClientType
	GetLatestHeight() uint64
	IsFrozen() bool
	GetFrozenHeight() uint64
	Validate() error
	GetProofSpecs() []*ics23.ProofSpec

	// Update and Misbehaviour functions

	CheckHeaderAndUpdateState(sdk.Context, codec.BinaryMarshaler, sdk.KVStore, Header) (ClientState, ConsensusState, error)
	CheckMisbehaviourAndUpdateState(sdk.Context, codec.BinaryMarshaler, sdk.KVStore, Misbehaviour) (ClientState, error)

	// State verification functions

	VerifyClientState(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		root commitmentexported.Root,
		height uint64,
		prefix commitmentexported.Prefix,
		counterpartyClientIdentifier string,
		proof []byte,
		clientState ClientState,
	) error
	VerifyClientConsensusState(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		root commitmentexported.Root,
		height uint64,
		counterpartyClientIdentifier string,
		consensusHeight uint64,
		prefix commitmentexported.Prefix,
		proof []byte,
		consensusState ConsensusState,
	) error
	VerifyConnectionState(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		height uint64,
		prefix commitmentexported.Prefix,
		proof []byte,
		connectionID string,
		connectionEnd connectionexported.ConnectionI,
	) error
	VerifyChannelState(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		height uint64,
		prefix commitmentexported.Prefix,
		proof []byte,
		portID,
		channelID string,
		channel channelexported.ChannelI,
	) error
	VerifyPacketCommitment(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		height uint64,
		prefix commitmentexported.Prefix,
		proof []byte,
		portID,
		channelID string,
		sequence uint64,
		commitmentBytes []byte,
	) error
	VerifyPacketAcknowledgement(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		height uint64,
		prefix commitmentexported.Prefix,
		proof []byte,
		portID,
		channelID string,
		sequence uint64,
		acknowledgement []byte,
	) error
	VerifyPacketAcknowledgementAbsence(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		height uint64,
		prefix commitmentexported.Prefix,
		proof []byte,
		portID,
		channelID string,
		sequence uint64,
	) error
	VerifyNextSequenceRecv(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		height uint64,
		prefix commitmentexported.Prefix,
		proof []byte,
		portID,
		channelID string,
		nextSequenceRecv uint64,
	) error
}

// ConsensusState is the state of the consensus process
type ConsensusState interface {
	ClientType() ClientType // Consensus kind

	// GetHeight returns the height of the consensus state
	GetHeight() uint64

	// GetRoot returns the commitment root of the consensus state,
	// which is used for key-value pair verification.
	GetRoot() commitmentexported.Root

	// GetTimestamp returns the timestamp (in nanoseconds) of the consensus state
	GetTimestamp() uint64

	ValidateBasic() error
}

// Misbehaviour defines counterparty misbehaviour for a specific consensus type
type Misbehaviour interface {
	evidenceexported.Evidence
	ClientType() ClientType
	GetClientID() string
}

// Header is the consensus state update information
type Header interface {
	ClientType() ClientType
	GetHeight() uint64
}

// message and evidence types for the IBC client
const (
	TypeMsgCreateClient             string = "create_client"
	TypeMsgUpdateClient             string = "update_client"
	TypeMsgSubmitClientMisbehaviour string = "submit_client_misbehaviour"
	TypeEvidenceClientMisbehaviour  string = "client_misbehaviour"
)

// MsgCreateClient defines the msg interface that the
// CreateClient Handler expects
type MsgCreateClient interface {
	sdk.Msg
	GetClientID() string
	GetClientType() string
	GetConsensusState() ConsensusState
	InitializeClientState() ClientState
}

// MsgUpdateClient defines the msg interface that the
// UpdateClient Handler expects
type MsgUpdateClient interface {
	sdk.Msg
	GetClientID() string
	GetHeader() Header
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
}

// ClientType defines the type of the consensus algorithm
type ClientType byte

// available client types
const (
	SoloMachine ClientType = 6
	Tendermint  ClientType = 7
	Localhost   ClientType = 9
)

// string representation of the client types
const (
	ClientTypeSoloMachine string = "solomachine"
	ClientTypeTendermint  string = "tendermint"
	ClientTypeLocalHost   string = "localhost"
)

func (ct ClientType) String() string {
	switch ct {
	case Tendermint:
		return ClientTypeTendermint
	case Localhost:
		return ClientTypeLocalHost
	default:
		return ""
	}
}

// MarshalJSON marshal to JSON using string.
func (ct ClientType) MarshalJSON() ([]byte, error) {
	return json.Marshal(ct.String())
}

// UnmarshalJSON decodes from JSON.
func (ct *ClientType) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	clientType := ClientTypeFromString(s)
	if clientType == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "invalid client type '%s'", s)
	}

	*ct = clientType
	return nil
}

// ClientTypeFromString returns a byte that corresponds to the registered client
// type. It returns 0 if the type is not found/registered.
func ClientTypeFromString(clientType string) ClientType {
	switch clientType {
	case ClientTypeTendermint:
		return Tendermint
	case ClientTypeLocalHost:
		return Localhost
	default:
		return 0
	}
}
