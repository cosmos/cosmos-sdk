package simapp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

var _ gov.MsgSubmitProposal = MsgGovSubmitProposal{}
var _ gov.Content = &SomeOtherProposal{}

func (m SomeOtherProposal) ProposalRoute() string {
	panic("implement me")
}

func (m SomeOtherProposal) ProposalType() string {
	panic("implement me")
}

func (m SomeOtherProposal) ValidateBasic() sdk.Error {
	return nil
}

func (m MsgGovSubmitProposal) GetContent() types.Content {
	if content := m.GetTextProposal(); content != nil {
		return content
	} else if content := m.GetSomeOtherProposal(); content != nil {
		return content
	}
	return nil
}



