package v038

import (
	"time"

	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
)

// DONTCOVER
// nolint

// Default parameter values
const (
	ModuleName            = "evidence"
	DefaultParamspace     = ModuleName
	DefaultMaxEvidenceAge = 60 * 2 * time.Second
)

// Params defines the total set of parameters for the evidence module
type Params struct {
	MaxEvidenceAge time.Duration `json:"max_evidence_age" yaml:"max_evidence_age"`
}

// GenesisState defines the evidence module's genesis state.
type GenesisState struct {
	Params   Params              `json:"params" yaml:"params"`
	Evidence []exported.Evidence `json:"evidence" yaml:"evidence"`
}
