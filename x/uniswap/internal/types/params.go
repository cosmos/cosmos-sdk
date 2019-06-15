package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// TODO: come up with better naming for these params?
// Parameter store keys
var (
	ParamStoreKeyFeeParams         = []byte("feeparams")
	ParamStoreKeyNativeAssetParams = []byte("nativeAssetparam")
)

// ParamKeyTable creates a ParamTable for uniswap module.
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable(
		ParamStoreKeyNativeAssetParams, NativeAssetParams{},
		ParamStoreKeyFeeParams, FeeParams{},
	)
}

// NativeAssetParams represents the native asset used in the uniswap module.
type NativeAssetParams struct {
	Denom string
}

// NewNativeAssetParams creates a new NativeAssetParams object
func NewNativeAssetParams(denom string) NativeAssetParams {
	return NativeAssetParams{
		Denom: denom,
	}
}

func (p NativeAssetParams) String() string {
	return fmt.Sprintf("Native Asset:	%s\n", p.Denom)
}

// FeeParams represents the numerator and denominator for calculating
// the swap fee.
// Fee = 1 - (FeeN / FeeD)
type FeeParams struct {
	FeeN sdk.Int `json:"fee_numerator"`   // fee numerator
	FeeD sdk.Int `json:"fee_denominator"` // fee denominator
}

// NewFeeParams creates a new FeeParams object
func NewFeeParams(feeN, feeD sdk.Int) FeeParams {
	return FeeParams{
		FeeN: feeN,
		FeeD: feeD,
	}
}

func (p FeeParams) String() string {
	return fmt.Sprintf(`Fee Params:
  Fee Numerator:	%s
  Fee Denominator:	%s
`,
		p.FeeN, p.FeeD,
	)
}

type Params struct {
	NativeAssetParams NativeAssetParams `json:"native_asset"`
	FeeParams         FeeParams         `json:"fee"`
}

// DefaultParams returns the default uniswap module parameters
func DefaultParams() Params {
	return Params{
		FeeParams: FeeParams{
			FeeN: sdk.NewInt(997),
			FeeD: sdk.NewInt(1000),
		},
		NativeAssetParams: NativeAssetParams{
			Denom: sdk.DefaultBondDenom,
		},
	}
}

// ValidateParams validates the set params
func ValidateParams(p Params) error {
	if p.FeeParams.FeeN.GT(p.FeeParams.FeeD) {
		return fmt.Errorf("fee numerator must be less than or equal to fee denominator")
	}
	if p.FeeParams.FeeD.Mod(sdk.NewInt(10)) != sdk.NewInt(0) {
		return fmt.Errorf("fee denominator must be multiple of 10")
	}
	if strings.TrimSpace(p.NativeAssetParams.Denom) != "" {
		return fmt.Errorf("native asset denomination must not be empty")
	}
	return nil
}
