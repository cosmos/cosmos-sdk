package types

import (
	"time"

	"github.com/cosmos/cosmos-sdk/x/params"

	"gopkg.in/yaml.v2"
)

// DONTCOVER

// Default parameter values
const (
	DefaultParamspace     = ModuleName
	DefaultMaxEvidenceAge = 60 * 2 * time.Second
)

// Parameter store keys
var (
	KeyMaxEvidenceAge = []byte("MaxEvidenceAge")

	// The Double Sign Jail period ends at Max Time supported by Amino
	// (Dec 31, 9999 - 23:59:59 GMT).
	DoubleSignJailEndTime = time.Unix(253402300799, 0)
)

// Params defines the total set of parameters for the evidence module
type Params struct {
	MaxEvidenceAge time.Duration `json:"max_evidence_age" yaml:"max_evidence_age"`
}

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

func (p Params) MarshalYAML() (interface{}, error) {
	bz, err := yaml.Marshal(p)
	return string(bz), err
}

func (p Params) String() string {
	out, _ := p.MarshalYAML()
	return out.(string)
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyMaxEvidenceAge, &p.MaxEvidenceAge),
	}
}

// DefaultParams returns the default parameters for the evidence module.
func DefaultParams() Params {
	return Params{
		MaxEvidenceAge: DefaultMaxEvidenceAge,
	}
}
