package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Deposit
type Deposit struct {
	ProposalID uint64         `json:"proposal_id" yaml:"proposal_id"` //  proposalID of the proposal
	Depositor  sdk.AccAddress `json:"depositor" yaml:"depositor"`     //  Address of the depositor
	Amount     sdk.Coins      `json:"amount" yaml:"amount"`           //  Deposit amount
}

// NewDeposit creates a new Deposit instance
func NewDeposit(proposalID uint64, depositor sdk.AccAddress, amount sdk.Coins) Deposit {
	return Deposit{proposalID, depositor, amount}
}

func (d Deposit) String() string {
	return fmt.Sprintf("deposit by %s on Proposal %d is for the amount %s",
		d.Depositor, d.ProposalID, d.Amount)
}

// Deposits is a collection of Deposit objects
type Deposits []Deposit

func (d Deposits) String() string {
	if len(d) == 0 {
		return "[]"
	}
	out := fmt.Sprintf("Deposits for Proposal %d:", d[0].ProposalID)
	for _, dep := range d {
		out += fmt.Sprintf("\n  %s: %s", dep.Depositor, dep.Amount)
	}
	return out
}

// Equals returns whether two deposits are equal.
func (d Deposit) Equals(comp Deposit) bool {
	return d.Depositor.Equals(comp.Depositor) && d.ProposalID == comp.ProposalID && d.Amount.IsEqual(comp.Amount)
}

// Empty returns whether a deposit is empty.
func (d Deposit) Empty() bool {
	return d.Equals(Deposit{})
}
