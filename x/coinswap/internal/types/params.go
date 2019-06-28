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

// Params defines the fee and native denomination for coinswap
type Params struct {
	NativeDenom string   `json:"native_denom"`
	Fee         FeeParam `json:"fee"`
}

func NewParams(nativeDenom string, fee FeeParam) Params {
	return Params{
		NativeDenom: nativeDenom,
		Fee:         fee,
	}
}

// FeeParam defines the numerator and denominator used in calculating the
// amount to be reserved as a liquidity fee.
// TODO: come up with a more descriptive name than Numerator/Denominator
// Fee = 1 - (Numerator / Denominator) TODO: move this to spec
type FeeParam struct {
	Numerator   sdk.Int `json:"fee_numerator"`
	Denominator sdk.Int `json:"fee_denominator"`
}

func NewFeeParam(numerator, denominator sdk.Int) FeeParam {
	return FeeParam{
		Numerator:   numerator,
		Denominator: denominator,
	}
}

// ParamKeyTable returns the KeyTable for coinswap module
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() params.ParamSetPairs {
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

// DefaultParams returns the default coinswap module parameters
func DefaultParams() Params {
	feeParam := NewFeeParam(sdk.NewInt(997), sdk.NewInt(1000))

	return Params{
		NativeDenom: sdk.DefaultBondDenom,
		Fee:         feeParam,
	}
}

// ValidateParams validates a set of params
func ValidateParams(p Params) error {
	// TODO: ensure equivalent sdk.validateDenom validation
	if strings.TrimSpace(p.NativeDenom) != "" {
		return fmt.Errorf("native denomination must not be empty")
	}
	if !p.Fee.Numerator.IsPositive() {
		return fmt.Errorf("fee numerator is not positive: %v", p.Fee.Numerator)
	}
	if !p.Fee.Denominator.IsPositive() {
		return fmt.Errorf("fee denominator is not positive: %v", p.Fee.Denominator)
	}
	if p.Fee.Numerator.GTE(p.Fee.Denominator) {
		return fmt.Errorf("fee numerator is greater than or equal to fee numerator")
	}
	return nil
}
