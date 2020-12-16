package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

const (
	// ProposalTypeClientUpdate defines the type for a ClientUpdateProposal
	ProposalTypeClientUpdate = "ClientUpdate"
)

var (
	_ govtypes.Content                   = &ClientUpdateProposal{}
	_ codectypes.UnpackInterfacesMessage = ClientUpdateProposal{}
)

// NewClientUpdateProposal creates a new client update proposal.
func NewClientUpdateProposal(title, description, clientID string, header exported.Header) (*ClientUpdateProposal, error) {
	any, err := PackHeader(header)
	if err != nil {
		return nil, err
	}

	return &ClientUpdateProposal{
		Title:       title,
		Description: description,
		ClientId:    clientID,
		Header:      any,
	}, nil
}

// GetTitle returns the title of a client update proposal.
func (cup *ClientUpdateProposal) GetTitle() string { return cup.Title }

// GetDescription returns the description of a client update proposal.
func (cup *ClientUpdateProposal) GetDescription() string { return cup.Description }

// ProposalRoute returns the routing key of a client update proposal.
func (cup *ClientUpdateProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a client update proposal.
func (cup *ClientUpdateProposal) ProposalType() string { return ProposalTypeClientUpdate }

// ValidateBasic runs basic stateless validity checks
func (cup *ClientUpdateProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(cup)
	if err != nil {
		return err
	}

	if err := host.ClientIdentifierValidator(cup.ClientId); err != nil {
		return err
	}

	header, err := UnpackHeader(cup.Header)
	if err != nil {
		return err
	}

	return header.ValidateBasic()
}

// UnpackInterfaces implements the UnpackInterfacesMessage interface.
func (cup ClientUpdateProposal) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var header exported.Header
	return unpacker.UnpackAny(cup.Header, &header)
}
