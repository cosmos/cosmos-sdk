package v040

import (
	"strings"

	"gopkg.in/yaml.v2"
)

var (
	BondStatusUnspecified = "BOND_STATUS_UNSPECIFIED"
	BondStatusUnbonded    = "BOND_STATUS_UNBONDED"
	BondStatusUnbonding   = "BOND_STATUS_UNBONDING"
	BondStatusBonded      = "BOND_STATUS_BONDED"
)

// String implements the Stringer interface for a Validator object.
func (v Validator) String() string {
	out, _ := yaml.Marshal(v)
	return string(out)
}

// Validators is a collection of Validator
type Validators []Validator

func (v Validators) String() (out string) {
	for _, val := range v {
		out += val.String() + "\n"
	}

	return strings.TrimSpace(out)
}

// ValidatorsByVotingPower implements sort.Interface for []Validator based on
// the VotingPower and Address fields.
// The validators are sorted first by their voting power (descending). Secondary index - Address (ascending).
// Copied from tendermint/types/validator_set.go
type ValidatorsByVotingPower []Validator

// String implements the Stringer interface for a Description object.
func (d Description) String() string {
	out, _ := yaml.Marshal(d)
	return string(out)
}
