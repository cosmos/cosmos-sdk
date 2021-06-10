package v040

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

var _ codectypes.UnpackInterfacesMessage = Proposal{}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (p Proposal) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var content types.Content
	return unpacker.UnpackAny(p.Content, &content)
}

// Proposals is an array of proposal
type Proposals []Proposal

var _ codectypes.UnpackInterfacesMessage = Proposals{}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (p Proposals) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, x := range p {
		err := x.UnpackInterfaces(unpacker)
		if err != nil {
			return err
		}
	}
	return nil
}

// Deposits is a collection of Deposit objects
type Deposits []Deposit

// Implements Content Interface
var _ types.Content = &TextProposal{}

// ProposalRoute returns the proposal router key
func (tp *TextProposal) ProposalRoute() string { return RouterKey }

// ProposalType is "Text"
func (tp *TextProposal) ProposalType() string { return types.ProposalTypeText }

// ValidateBasic validates the content's title and description of the proposal
func (tp *TextProposal) ValidateBasic() error { return types.ValidateAbstract(tp) }
