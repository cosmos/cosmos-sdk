package types

import (
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

const (
	// ProposalTypeClientUpdate defines the type for a ClientUpdateProposal
	ProposalTypeClientUpdate = "ClientUpdate"
)

// Assert ClientUpdateProposal implements govtypes.Content at compile-time
var _ govtypes.Content = &ClientUpdateProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeClientUpdate)
	govtypes.RegisterProposalTypeCodec(&ClientUpdateProposal{}, "cosmos-sdk/ClientUpdateProposal")
}

// NewClientUpdateProposal creates a new client update proposal.
func NewClientUpdateProposal(title, description, clientID string, header exported.Header) (*ClientUpdateProposal, error) {
	any, err := PackHeader(header)
	if err != nil {
		return nil, err
	}

	return &ClientUpdateProposal{title, description, clientID, any}, nil
}

// GetTitle returns the title of a community pool spend proposal.
func (cup *ClientUpdateProposal) GetTitle() string { return cup.Title }

// GetDescription returns the description of a community pool spend proposal.
func (cup *ClientUpdateProposal) GetDescription() string { return cup.Description }

// GetDescription returns the routing key of a community pool spend proposal.
func (cup *ClientUpdateProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool spend proposal.
func (cup *ClientUpdateProposal) ProposalType() string { return ProposalTypeClientUpdate }

// ValidateBasic runs basic stateless validity checks
func (cup *ClientUpdateProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(cup)
	if err != nil {
		return err
	}

	return nil
}
