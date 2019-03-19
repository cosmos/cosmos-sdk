package params

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/proposal"
)

// Router key for params module
const RouterKey = "params"

// MsgSubmitParameterChangeProposal submits a proposal to change multiple params
type MsgSubmitParameterChangeProposal struct {
	Title          string         `json:"title"`           //  Title of the proposal
	Description    string         `json:"description"`     //  Description of the proposal
	Proposer       sdk.AccAddress `json:"proposer"`        //  Address of the proposer
	InitialDeposit sdk.Coins      `json:"initial_deposit"` //  Initial deposit paid by sender. Must be strictly positive.
	Changes        []Change       `json:"changes"`         // Parameters to be changed
}

// Constructs new MsgSubmitParameterChangeProposal
func NewMsgSubmitParameterChangeProposal(title, description string, changes []Change, proposer sdk.AccAddress, initialDeposit sdk.Coins) MsgSubmitParameterChangeProposal {
	return MsgSubmitParameterChangeProposal{
		Title:          title,
		Description:    description,
		Proposer:       proposer,
		InitialDeposit: initialDeposit,
		Changes:        changes,
	}
}

var _ sdk.Msg = MsgSubmitParameterChangeProposal{}

// Implements sdk.Msg
func (msg MsgSubmitParameterChangeProposal) Route() string {
	return RouterKey
}

// Implements sdk.Msg
func (msg MsgSubmitParameterChangeProposal) Type() string {
	return "submit_parameter_change_proposal"
}

// Implements sdk.Msg
func (msg MsgSubmitParameterChangeProposal) ValidateBasic() sdk.Error {
	err := proposal.ValidateMsgBasic(msg.Title, msg.Description, msg.Proposer, msg.InitialDeposit)
	if err != nil {
		return err
	}
	return ValidateChanges(msg.Changes)
}

// Implements sdk.Msg
func (msg MsgSubmitParameterChangeProposal) GetSignBytes() []byte {
	bz := msgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// Implements sdk.Msg
func (msg MsgSubmitParameterChangeProposal) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Proposer}
}
