package exported

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// ClientState defines the required common functions for light clients.
type ClientState interface {
	GetID() string
	ClientType() ClientType
	GetLatestHeight() uint64
	IsFrozen() bool

	// State verification functions

	VerifyClientConsensusState(
		cdc *codec.Codec,
		height uint64,
		prefix commitment.PrefixI,
		proof commitment.ProofI,
		consensusState ConsensusState,
	) error
	VerifyConnectionState(
		cdc *codec.Codec,
		height uint64,
		prefix commitment.PrefixI,
		proof commitment.ProofI,
		connectionID string,
		connectionEnd connectionexported.ConnectionI,
		consensusState ConsensusState,
	) error
	VerifyChannelState(
		cdc *codec.Codec,
		height uint64,
		prefix commitment.PrefixI,
		proof commitment.ProofI,
		portID,
		channelID string,
		channel channelexported.ChannelI,
		consensusState ConsensusState,
	) error
	VerifyPacketCommitment(
		height uint64,
		prefix commitment.PrefixI,
		proof commitment.ProofI,
		portID,
		channelID string,
		sequence uint64,
		commitmentBytes []byte,
		consensusState ConsensusState,
	) error
	VerifyPacketAcknowledgement(
		height uint64,
		prefix commitment.PrefixI,
		proof commitment.ProofI,
		portID,
		channelID string,
		sequence uint64,
		acknowledgement []byte,
		consensusState ConsensusState,
	) error
	VerifyPacketAcknowledgementAbsence(
		height uint64,
		prefix commitment.PrefixI,
		proof commitment.ProofI,
		portID,
		channelID string,
		sequence uint64,
		consensusState ConsensusState,
	) error
	VerifyNextSequenceRecv(
		height uint64,
		prefix commitment.PrefixI,
		proof commitment.ProofI,
		portID,
		channelID string,
		nextSequenceRecv uint64,
		consensusState ConsensusState,
	) error
}

// ConsensusState is the state of the consensus process
type ConsensusState interface {
	ClientType() ClientType // Consensus kind

	// GetRoot returns the commitment root of the consensus state,
	// which is used for key-value pair verification.
	GetRoot() commitment.RootI

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

// ClientType defines the type of the consensus algorithm
type ClientType byte

// available client types
const (
	Tendermint ClientType = iota + 1 // 1
)

// string representation of the client types
const (
	ClientTypeTendermint string = "tendermint"
)

func (ct ClientType) String() string {
	switch ct {
	case Tendermint:
		return ClientTypeTendermint
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

	default:
		return 0
	}
}
