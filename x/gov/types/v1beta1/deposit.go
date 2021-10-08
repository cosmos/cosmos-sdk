package v1beta1

import (
	"fmt"

	"sigs.k8s.io/yaml"
)

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

// Returns a string representation of the deposit
func (d Deposit) String() string {
	out, _ := yaml.Marshal(d)
	return string(out)
}

// Empty returns whether a deposit is empty.
func (d Deposit) Empty() bool {
	return d.String() == Deposit{}.String()
}
