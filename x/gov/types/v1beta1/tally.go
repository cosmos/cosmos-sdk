package v1beta1

import (
	"sigs.k8s.io/yaml"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewTallyResult creates a new TallyResult instance
func NewTallyResult(yes, abstain, no, noWithVeto sdk.Int) TallyResult {
	return TallyResult{
		Yes:        yes,
		Abstain:    abstain,
		No:         no,
		NoWithVeto: noWithVeto,
	}
}

// EmptyTallyResult returns an empty TallyResult.
func EmptyTallyResult() TallyResult {
	return NewTallyResult(sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt())
}

// Equals returns if two proposals are equal.
func (tr TallyResult) Equals(comp TallyResult) bool {
	return tr.Yes.Equal(comp.Yes) &&
		tr.Abstain.Equal(comp.Abstain) &&
		tr.No.Equal(comp.No) &&
		tr.NoWithVeto.Equal(comp.NoWithVeto)
}

// String implements stringer interface
func (tr TallyResult) String() string {
	out, _ := yaml.Marshal(tr)
	return string(out)
}
