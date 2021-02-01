package v040

import (
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// String returns a human readable string representation of a Delegation.
func (d Delegation) String() string {
	out, _ := yaml.Marshal(d)
	return string(out)
}

// Delegations is a collection of delegations
type Delegations []Delegation

func (d Delegations) String() (out string) {
	for _, del := range d {
		out += del.String() + "\n"
	}

	return strings.TrimSpace(out)
}

// String implements the stringer interface for a UnbondingDelegationEntry.
func (e UnbondingDelegationEntry) String() string {
	out, _ := yaml.Marshal(e)
	return string(out)
}

// String returns a human readable string representation of an UnbondingDelegation.
func (ubd UnbondingDelegation) String() string {
	out := fmt.Sprintf(`Unbonding Delegations between:
  Delegator:                 %s
  Validator:                 %s
	Entries:`, ubd.DelegatorAddress, ubd.ValidatorAddress)
	for i, entry := range ubd.Entries {
		out += fmt.Sprintf(`    Unbonding Delegation %d:
      Creation Height:           %v
      Min time to unbond (unix): %v
      Expected balance:          %s`, i, entry.CreationHeight,
			entry.CompletionTime, entry.Balance)
	}

	return out
}

// UnbondingDelegations is a collection of UnbondingDelegation
type UnbondingDelegations []UnbondingDelegation

func (ubds UnbondingDelegations) String() (out string) {
	for _, u := range ubds {
		out += u.String() + "\n"
	}

	return strings.TrimSpace(out)
}

// String implements the Stringer interface for a RedelegationEntry object.
func (e RedelegationEntry) String() string {
	out, _ := yaml.Marshal(e)
	return string(out)
}

// String returns a human readable string representation of a Redelegation.
func (red Redelegation) String() string {
	out := fmt.Sprintf(`Redelegations between:
  Delegator:                 %s
  Source Validator:          %s
  Destination Validator:     %s
  Entries:
`,
		red.DelegatorAddress, red.ValidatorSrcAddress, red.ValidatorDstAddress,
	)

	for i, entry := range red.Entries {
		out += fmt.Sprintf(`    Redelegation Entry #%d:
      Creation height:           %v
      Min time to unbond (unix): %v
      Dest Shares:               %s
`,
			i, entry.CreationHeight, entry.CompletionTime, entry.SharesDst,
		)
	}

	return strings.TrimRight(out, "\n")
}

// Redelegations are a collection of Redelegation
type Redelegations []Redelegation

func (d Redelegations) String() (out string) {
	for _, red := range d {
		out += red.String() + "\n"
	}

	return strings.TrimSpace(out)
}
