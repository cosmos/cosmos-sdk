package types

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewDeposit creates a new Deposit instance
//nolint:interfacer
func NewDeposit(proposalID uint64, depositor sdk.AccAddress, amount sdk.Coins) Deposit {
	return Deposit{proposalID, depositor.String(), amount}
}

func (d Deposit) String() string {
	out, _ := yaml.Marshal(d)
	return string(out)
}

// Deposits is a collection of Deposit objects
type Deposits []Deposit

// Equal returns true if two slices (order-dependant) of deposits are equal.
func (d Deposits) Equal(other Deposits) bool {
	if len(d) != len(other) {
		return false
	}

	for i, deposit := range d {
		if deposit.String() != other[i].String() {
			return false
		}
	}

	return true
}

func (d Deposits) String() string {
	if len(d) == 0 {
		return "[]"
	}
	out := fmt.Sprintf("Deposits for Proposal %d:", d[0].ProposalId)
	for _, dep := range d {
		out += fmt.Sprintf("\n  %s: %s", dep.Depositor, dep.Amount)
	}
	return out
}

// Empty returns whether a deposit is empty.
func (d Deposit) Empty() bool {
	return d.String() == Deposit{}.String()
}
