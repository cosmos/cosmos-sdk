package proposal

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/errors"
)

// nolint
const (
	ModuleName = "gov"

	RouterKey = ModuleName
)

// SubmitForm
type SubmitForm struct {
	Title          string         `json:"title"`           //  Title of the proposal
	Description    string         `json:"description"`     //  Description of the proposal
	Proposer       sdk.AccAddress `json:"proposer"`        //  Address of the proposer
	InitialDeposit sdk.Coins      `json:"initial_deposit"` //  Initial deposit paid by sender. Must be strictly positive.
}

func NewSubmitForm(title, description string, proposer sdk.AccAddress, initialDeposit sdk.Coins) SubmitForm {
	return SubmitForm{
		Title:          title,
		Description:    description,
		Proposer:       proposer,
		InitialDeposit: initialDeposit,
	}
}

// Partially implements sdk.Msg
func (form SubmitForm) ValidateBasic() sdk.Error {
	// XXX: we are already checking IsValidContent in Submit.
	// Is it efficient to put it in ValidateBasic?
	err := IsValidContent(errors.DefaultCodespace, form.Title, form.Description)
	if err != nil {
		return err
	}
	if form.Proposer.Empty() {
		return sdk.ErrInvalidAddress(form.Proposer.String())
	}
	if !form.InitialDeposit.IsValid() {
		return sdk.ErrInvalidCoins(form.InitialDeposit.String())
	}
	if form.InitialDeposit.IsAnyNegative() {
		return sdk.ErrInvalidCoins(form.InitialDeposit.String())
	}
	return nil
}

// Partially implemets sdk.Msg
func (form SubmitForm) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{form.Proposer}
}
