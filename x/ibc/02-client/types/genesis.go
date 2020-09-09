package types

import (
	"fmt"
	"sort"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
)

var (
	_ codectypes.UnpackInterfacesMessage = IdentifiedClientState{}
	_ codectypes.UnpackInterfacesMessage = ClientsConsensusStates{}
	_ codectypes.UnpackInterfacesMessage = ClientConsensusStates{}
	_ codectypes.UnpackInterfacesMessage = GenesisState{}
)

var _ sort.Interface = ClientsConsensusStates{}

// ClientsConsensusStates defines a slice of ClientConsensusStates that supports the sort interface
type ClientsConsensusStates []ClientConsensusStates

// Len implements sort.Interface
func (ccs ClientsConsensusStates) Len() int { return len(ccs) }

// Less implements sort.Interface
func (ccs ClientsConsensusStates) Less(i, j int) bool { return ccs[i].ClientId < ccs[j].ClientId }

// Swap implements sort.Interface
func (ccs ClientsConsensusStates) Swap(i, j int) { ccs[i], ccs[j] = ccs[j], ccs[i] }

// Sort is a helper function to sort the set of ClientsConsensusStates in place
func (ccs ClientsConsensusStates) Sort() ClientsConsensusStates {
	sort.Sort(ccs)
	return ccs
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (ccs ClientsConsensusStates) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, clientConsensus := range ccs {
		if err := clientConsensus.UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}
	return nil
}

// NewClientConsensusStates creates a new ClientConsensusStates instance.
func NewClientConsensusStates(clientID string, consensusStates []ConsensusStateWithHeight) ClientConsensusStates {
	return ClientConsensusStates{
		ClientId:        clientID,
		ConsensusStates: consensusStates,
	}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (ccs ClientConsensusStates) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, consStateWithHeight := range ccs.ConsensusStates {
		var consensusState exported.ConsensusState
		err := unpacker.UnpackAny(consStateWithHeight.ConsensusState, &consensusState)
		if err != nil {
			return err
		}
	}
	return nil
}

// NewGenesisState creates a GenesisState instance.
func NewGenesisState(
	clients []IdentifiedClientState, clientsConsensus ClientsConsensusStates, createLocalhost bool,
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
		Clients:          []IdentifiedClientState{},
		ClientsConsensus: ClientsConsensusStates{},
		CreateLocalhost:  true,
	}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (gs GenesisState) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, client := range gs.Clients {
		if err := client.UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}

	return gs.ClientsConsensus.UnpackInterfaces(unpacker)
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	for i, client := range gs.Clients {
		if err := host.ClientIdentifierValidator(client.ClientId); err != nil {
			return fmt.Errorf("invalid client consensus state identifier %s index %d: %w", client.ClientId, i, err)
		}

		clientState, ok := client.ClientState.GetCachedValue().(exported.ClientState)
		if !ok {
			return fmt.Errorf("invalid client state")
		}
		if err := clientState.Validate(); err != nil {
			return fmt.Errorf("invalid client %v index %d: %w", client, i, err)
		}
	}

	for i, cc := range gs.ClientsConsensus {
		if err := host.ClientIdentifierValidator(cc.ClientId); err != nil {
			return fmt.Errorf("invalid client consensus state identifier %s index %d: %w", cc.ClientId, i, err)
		}

		for _, consensusState := range cc.ConsensusStates {
			if consensusState.Height.IsZero() {
				return fmt.Errorf("consensus state height cannot be zero")
			}

			cs, ok := consensusState.ConsensusState.GetCachedValue().(exported.ConsensusState)
			if !ok {
				return fmt.Errorf("invalid consensus state")
			}

			if err := cs.ValidateBasic(); err != nil {
				return fmt.Errorf("invalid client consensus state %v index %d: %w", cs, i, err)
			}
		}
	}

	return nil
}
