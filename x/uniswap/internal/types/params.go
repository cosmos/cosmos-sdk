package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// TODO: come up with better naming for these params?
// Parameter store keys
var (
	ParamStoreKeyFeeParams = []byte("feeparams")
)

// uniswap parameters
// Fee = 1 - (FeeN - FeeD)
type FeeParams struct {
	FeeN sdk.Int `json:"fee_numerator"`   // fee numerator
	FeeD sdk.Int `json:"fee_denominator"` // fee denominator
}

// ParamKeyTable creates a ParamTable for uniswap module.
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable(
		ParamStoreKeyFeeParams, FeeParams{},
	)
}

// NewFeeParams is a constructor for the FeeParams struct
func NewFeeParams(feeN, feeD sdk.Int) FeeParams {
	return FeeParams{
		FeeN: feeN,
		FeeD: feeD,
	}
}

// DefaultParams returns the default uniswap module parameters
func DefaultParams() FeeParams {
	return FeeParams{
		FeeN: sdk.NewInt(997),
		FeeD: sdk.NewInt(1000),
	}
}

// ValidateParams validates the set params
func ValidateParams(p FeeParams) error {
	if p.FeeN.GT(p.FeeD) {
		return fmt.Errorf("fee numerator must be less than or equal to fee denominator")
	}
	if p.FeeD.Mod(sdk.NewInt(10)) != sdk.NewInt(0) {
		return fmt.Errorf("fee denominator must be multiple of 10")
	}
	return nil
}

func (p FeeParams) String() string {
	return fmt.Sprintf(`Uniswap Params:
  Fee Numerator:	%s
  Fee Denominator:	%s
`,
		p.FeeN, p.FeeD,
	)
}
