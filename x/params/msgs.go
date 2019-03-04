package params

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MsgSubmitParameterChangeProposal struct {
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	Changes        []Change       `json:"changes"`
	Proposer       sdk.AccAddress `json:"proposer"`
	InitialDeposit sdk.Coins      `json:"initial_deposit"`
}

func NewMsgSubmitProposal(title, description string, changes []Change, proposer sdk.AccAddress, initialDeposit sdk.Coins) MsgSubmitParameterChangeProposal {
	return MsgSubmitParameterChangeProposal{
		Title:          title,
		Description:    description,
		Changes:        changes,
		Proposer:       proposer,
		InitialDeposit: initialDeposit,
	}
}
