package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

const (
	DefaultParamspace = ModuleName
)

// Parameter store keys
var (
	KeyNativeDenom = []byte("nativeDenom")
	KeyFee         = []byte("fee")
)

// Params defines the fee and native denomination for uniswap
type Params struct {
	NativeDenom string  `json:"native_denom"`
	Fee         sdk.Dec `json:"fee"`
}

func NewParams(nativeDenom string, fee sdk.Dec) Params {

	return Params{
		NativeDenom: nativeDenom,
		Fee:         fee,
	}
}

// Implements params.ParamSet.
func (p *Params) ParamSetPair() params.ParamSetPairs {

	return params.ParamSetPairs{
		{KeyNativeDenom, &p.NativeDenom},
		{KeyFee, &p.Fee},
	}
}

// String returns a human readable string representation of the parameters.
func (p Params) String() string {
	return fmt.Sprintf(`Params:
Native Denom:	%s
Fee:			%s`, p.NativeDenom, p.Fee,
	)
}

// DefaultParams returns the default uniswap module parameters
func DefaultParams() Params {
	fee, err := sdk.NewDecFromStr("0.03")
	if err != nil {
		panic(err)
	}

	return Params{
		NativeDenom: sdk.DefaultBondDenom,
		Fee:         fee,
	}
}

// ValidateParams validates a set of params
func ValidateParams(p Params) error {
	// TODO: ensure equivalent sdk.validateDenom validation
	if strings.TrimSpace(p.NativeDenom) != "" {
		return fmt.Errorf("native denomination must not be empty")
	}
	if !p.Fee.IsPositive() {
		return fmt.Errorf("fee is not positive: %v", p.Fee)
	}
	if !p.Fee.LT(sdk.OneDec()) {
		return fmt.Errorf("fee must be less than one: %v", p.Fee)
	}
	return nil
}
