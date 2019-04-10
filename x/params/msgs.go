package params

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/proposal"
)

// Router key for params module
const RouterKey = "params"

// MsgSubmitProposal submits a proposal to change multiple params
type MsgSubmitProposal struct {
	Content        proposal.Content `json:"content"`
	Proposer       sdk.AccAddress   `json:"proposer"`        //  Address of the proposer
	InitialDeposit sdk.Coins        `json:"initial_deposit"` //  Initial deposit paid by sender. Must be strictly positive.
}

// Constructs new MsgSubmitProposal
func NewMsgSubmitProposal(title, description string, changes []Change, proposer sdk.AccAddress, initialDeposit sdk.Coins) MsgSubmitProposal {
	return MsgSubmitProposal{
		Content:        NewChangeProposal(title, description, changes),
		Proposer:       proposer,
		InitialDeposit: initialDeposit,
	}
}

var _ sdk.Msg = MsgSubmitProposal{}

// Implements sdk.Msg
func (msg MsgSubmitProposal) Route() string {
	return RouterKey
}

// Implements sdk.Msg
func (msg MsgSubmitProposal) Type() string {
	return "submit_parameter_change_proposal"
}

// Implements sdk.Msg
func (msg MsgSubmitProposal) ValidateBasic() sdk.Error {
	err := proposal.ValidateMsgBasic(msg.Proposer, msg.InitialDeposit)
	if err != nil {
		return err
	}
	return msg.Content.ValidateBasic()
}

// Implements sdk.Msg
func (msg MsgSubmitProposal) GetSignBytes() []byte {
	bz := msgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// Implements sdk.Msg
func (msg MsgSubmitProposal) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Proposer}
}
