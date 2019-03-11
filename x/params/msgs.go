package params

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const RouteKey = "params"

type MsgSubmitParameterChangeProposal struct {
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	Space          string         `json:"space"`
	Changes        []Change       `json:"changes"`
	Proposer       sdk.AccAddress `json:"proposer"`
	InitialDeposit sdk.Coins      `json:"initial_deposit"`
}

func NewMsgSubmitParameterChangeProposal(title, description string, changes []Change, proposer sdk.AccAddress, initialDeposit sdk.Coins) MsgSubmitParameterChangeProposal {
	return MsgSubmitParameterChangeProposal{
		Title:          title,
		Description:    description,
		Changes:        changes,
		Proposer:       proposer,
		InitialDeposit: initialDeposit,
	}
}

var _ sdk.Msg = MsgSubmitParameterChangeProposal{}

func (msg MsgSubmitParameterChangeProposal) Route() string {
	return RouteKey
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
