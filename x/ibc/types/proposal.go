package types

import (
	fmt "fmt"
	"strings"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
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
func NewClientUpdateProposal(title, description, clientID string, header types.Header) (*ClientUpdateProposal, error) {
	headerBytes, err := types.MarshalHeader(header)
	if err != nil {
		return &ClientUpdateProposal{}, err
	}

	return &ClientUpdateProposal{title, description, clientID, headerBytes}, nil
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

// String implements the Stringer interface.
func (cup ClientUpdateProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Client Update Proposal:
  Title:       %s
  Description: %s
  ClientID:   %s
  Header:      %s
`, cup.Title, cup.Description, cup.ClientId, cup.Header))
	return b.String()
}
