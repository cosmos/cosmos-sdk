package types

import (
	"fmt"

	proto "github.com/gogo/protobuf/proto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// NewGenesisClientState creates a new GenesisClientState instance.
func NewGenesisClientState(clientID string, clientState exported.ClientState) GenesisClientState {
	msg, ok := clientState.(proto.Message)
	if !ok {
		panic(fmt.Errorf("cannot proto marshal %T", clientState))
	}

	anyClientState, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		panic(err)
	}

	return GenesisClientState{
		ClientID:    clientID,
		ClientState: anyClientState,
	}
}

// ClientConsensusStates defines all the stored consensus states for a given client.
type ClientConsensusStates struct {
	ClientID        string                    `json:"client_id" yaml:"client_id"`
	ConsensusStates []exported.ConsensusState `json:"consensus_states" yaml:"consensus_states"`
}

// NewClientConsensusStates creates a new ClientConsensusStates instance.
func NewClientConsensusStates(id string, states []exported.ConsensusState) ClientConsensusStates {
	return ClientConsensusStates{
		ClientID:        id,
		ConsensusStates: states,
	}
}

// GenesisState defines the ibc client submodule's genesis state.
type GenesisState struct {
	Clients          []GenesisClientState    `json:"clients" yaml:"clients"`
	ClientsConsensus []ClientConsensusStates `json:"clients_consensus" yaml:"clients_consensus"`
	CreateLocalhost  bool                    `json:"create_localhost" yaml:"create_localhost"`
}

// NewGenesisState creates a GenesisState instance.
func NewGenesisState(
	clients []GenesisClientState, clientsConsensus []ClientConsensusStates, createLocalhost bool,
) GenesisState {
	return GenesisState{
		Clients:          clients,
		ClientsConsensus: clientsConsensus,
		CreateLocalhost:  createLocalhost,
	}
}

// DefaultGenesisState returns the ibc client submodule's default genesis state.
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Clients:          []GenesisClientState{},
		ClientsConsensus: []ClientConsensusStates{},
		CreateLocalhost:  true,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	for i, client := range gs.Clients {
		if err := host.ClientIdentifierValidator(client.ClientID); err != nil {
			return fmt.Errorf("invalid client consensus state identifier %s index %d: %w", client.ClientID, i, err)
		}
		cs, ok := client.ClientState.GetCachedValue().(exported.ClientState)
		if !ok {
			return fmt.Errorf("invalid client state")
		}
		if err := cs.Validate(); err != nil {
			return fmt.Errorf("invalid client %v index %d: %w", client, i, err)
		}
	}

	for i, cs := range gs.ClientsConsensus {
		if err := host.ClientIdentifierValidator(cs.ClientID); err != nil {
			return fmt.Errorf("invalid client consensus state identifier %s index %d: %w", cs.ClientID, i, err)
		}
		for _, consensusState := range cs.ConsensusStates {
			if err := consensusState.ValidateBasic(); err != nil {
				return fmt.Errorf("invalid client consensus state %v index %d: %w", consensusState, i, err)
			}
		}
	}

	return nil
}
