package params

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/proposal"
)

const RouterKey = "params"

type MsgSubmitParameterChangeProposal struct {
	proposal.SubmitForm `json:"submit_form"`
	Space               string   `json:"space"`
	Changes             []Change `json:"changes"`
}

func NewMsgSubmitParameterChangeProposal(title, description string, space string, changes []Change, proposer sdk.AccAddress, initialDeposit sdk.Coins) MsgSubmitParameterChangeProposal {
	return MsgSubmitParameterChangeProposal{
		SubmitForm: proposal.NewSubmitForm(title, description, proposer, initialDeposit),
		Space:      space,
		Changes:    changes,
	}
}

var _ sdk.Msg = MsgSubmitParameterChangeProposal{}

func (msg MsgSubmitParameterChangeProposal) Route() string {
	return RouterKey
}

func (msg MsgSubmitParameterChangeProposal) Type() string {
	return "submit_parameter_change_proposal"
}

func (msg MsgSubmitParameterChangeProposal) ValidateBasic() sdk.Error {
	// XXX
	return nil
}

func (msg MsgSubmitParameterChangeProposal) GetSignBytes() []byte {
	// XXX
	return nil
}

func (msg MsgSubmitParameterChangeProposal) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Proposer}
}
