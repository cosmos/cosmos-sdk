package exported

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/codec"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

// ClientState defines the required common functions for light clients.
type ClientState interface {
	GetID() string
	GetChainID() string
	ClientType() ClientType
	GetLatestHeight() uint64
	IsFrozen() bool
	Validate() error

	// State verification functions

	VerifyClientConsensusState(
		cdc *codec.Codec,
		root commitmentexported.Root,
		height uint64,
		counterpartyClientIdentifier string,
		consensusHeight uint64,
		prefix commitmentexported.Prefix,
		proof commitmentexported.Proof,
		consensusState ConsensusState,
	) error
	VerifyConnectionState(
		cdc codec.Marshaler,
		height uint64,
		prefix commitmentexported.Prefix,
		proof commitmentexported.Proof,
		connectionID string,
		connectionEnd connectionexported.ConnectionI,
		consensusState ConsensusState,
	) error
	VerifyChannelState(
		cdc codec.Marshaler,
		height uint64,
		prefix commitmentexported.Prefix,
		proof commitmentexported.Proof,
		portID,
		channelID string,
		channel channelexported.ChannelI,
		consensusState ConsensusState,
	) error
	VerifyPacketCommitment(
		height uint64,
		prefix commitmentexported.Prefix,
		proof commitmentexported.Proof,
		portID,
		channelID string,
		sequence uint64,
		commitmentBytes []byte,
		consensusState ConsensusState,
	) error
	VerifyPacketAcknowledgement(
		height uint64,
		prefix commitmentexported.Prefix,
		proof commitmentexported.Proof,
		portID,
		channelID string,
		sequence uint64,
		acknowledgement []byte,
		consensusState ConsensusState,
	) error
	VerifyPacketAcknowledgementAbsence(
		height uint64,
		prefix commitmentexported.Prefix,
		proof commitmentexported.Proof,
		portID,
		channelID string,
		sequence uint64,
		consensusState ConsensusState,
	) error
	VerifyNextSequenceRecv(
		height uint64,
		prefix commitmentexported.Prefix,
		proof commitmentexported.Proof,
		portID,
		channelID string,
		nextSequenceRecv uint64,
		consensusState ConsensusState,
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

// Misbehaviour defines a specific consensus kind and an evidence
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

// MsgCreateClient defines the msg interface that the
// CreateClient Handler expects
type MsgCreateClient interface {
	sdk.Msg
	GetClientID() string
	GetClientType() string
	GetConsensusState() ConsensusState
}

// MsgUpdateClient defines the msg interface that the
// UpdateClient Handler expects
type MsgUpdateClient interface {
	sdk.Msg
	GetClientID() string
	GetHeader() Header
}

// ClientType defines the type of the consensus algorithm
type ClientType byte

// available client types
const (
	Tendermint ClientType = iota + 1 // 1
	Localhost
)

// string representation of the client types
const (
	ClientTypeTendermint string = "tendermint"
	ClientTypeLocalHost  string = "localhost"
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
		return fmt.Errorf("invalid client type '%s'", s)
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
