package std

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var (
	_ gov.Codec = (*Codec)(nil)
)

// Codec defines the application-level codec. This codec contains all the
// required module-specific codecs that are to be provided upon initialization.
type Codec struct {
	codec.Marshaler

	// Keep reference to the amino codec to allow backwards compatibility along
	// with type, and interface registration.
	amino *codec.Codec

	anyUnpacker types.AnyUnpacker
}

func NewAppCodec(amino *codec.Codec, anyUnpacker types.AnyUnpacker) *Codec {
	return &Codec{Marshaler: codec.NewHybridCodec(amino, anyUnpacker), amino: amino, anyUnpacker: anyUnpacker}
}

// MarshalProposal marshals a Proposal. It accepts a Proposal defined by the x/gov
// module and uses the application-level Proposal type which has the concrete
// Content implementation to serialize.
func (c *Codec) MarshalProposal(p gov.Proposal) ([]byte, error) {
	proposal := &Proposal{ProposalBase: p.ProposalBase}
	if err := proposal.Content.SetContent(p.Content); err != nil {
		return nil, err
	}

	return c.Marshaler.MarshalBinaryBare(proposal)
}

// UnmarshalProposal decodes a Proposal defined by the x/gov module and uses the
// application-level Proposal type which has the concrete Content implementation
// to deserialize.
func (c *Codec) UnmarshalProposal(bz []byte) (gov.Proposal, error) {
	proposal := &Proposal{}
	if err := c.Marshaler.UnmarshalBinaryBare(bz, proposal); err != nil {
		return gov.Proposal{}, err
	}

	return gov.Proposal{
		Content:      proposal.Content.GetContent(),
		ProposalBase: proposal.ProposalBase,
	}, nil
}

// ----------------------------------------------------------------------------
// necessary types and interfaces registered. This codec is provided to all the
// modules the application depends on.
//
// NOTE: This codec will be deprecated in favor of AppCodec once all modules are
// migrated.
func MakeCodec(bm module.BasicManager) *codec.Codec {
	cdc := codec.New()

	bm.RegisterCodec(cdc)
	vesting.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	return cdc
}

// RegisterInterfaces registers Interfaces from sdk/types and vesting
func RegisterInterfaces(interfaceRegistry types.InterfaceRegistry) {
	sdk.RegisterInterfaces(interfaceRegistry)
	vesting.RegisterInterfaces(interfaceRegistry)
}
