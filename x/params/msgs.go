package params

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/proposal"
)

// Router key for params module
const RouterKey = "params"

// MsgSubmitParameterChangeProposal submits a proposal to change multiple params
type MsgSubmitParameterChangeProposal struct {
	proposal.SubmitForm `json:"submit_form"`
	Changes             []Change `json:"changes"`
}

// Constructs new MsgSubmitParameterChangeProposal
func NewMsgSubmitParameterChangeProposal(title, description string, changes []Change, proposer sdk.AccAddress, initialDeposit sdk.Coins) MsgSubmitParameterChangeProposal {
	return MsgSubmitParameterChangeProposal{
		SubmitForm: proposal.NewSubmitForm(title, description, proposer, initialDeposit),
		Changes:    changes,
	}
}

var _ sdk.Msg = MsgSubmitParameterChangeProposal{}

// Implements sdk.Msg
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
