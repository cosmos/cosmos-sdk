package types

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v2"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
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

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMaxEvidenceAge, &p.MaxEvidenceAge, validateMaxEvidenceAge),
	}
}

// DefaultParams returns the default parameters for the evidence module.
func DefaultParams() Params {
	return Params{
		MaxEvidenceAge: DefaultMaxEvidenceAge,
	}
}

func validateMaxEvidenceAge(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("max evidence age must be positive: %s", v)
	}

	return nil
}
