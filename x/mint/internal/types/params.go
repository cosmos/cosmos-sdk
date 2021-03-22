package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Parameter store keys
var (
	KeyMintDenom = []byte("MintDenom")
	//KeyInflationRateChange = []byte("InflationRateChange")
	//KeyInflationMax        = []byte("InflationMax")
	//KeyInflationMin        = []byte("InflationMin")
	//KeyGoalBonded          = []byte("GoalBonded")
	KeyBlocksPerYear = []byte("BlocksPerYear")

	KeyDeflationRate  = []byte("DeflationRate")
	KeyDeflationEpoch = []byte("DeflationEpoch")
	KeyFarmProportion = []byte("YieldFarmingProportion")
)

// mint parameters
type Params struct {
	MintDenom           string  `json:"mint_denom" yaml:"mint_denom"`                       // type of coin to mint
	InflationRateChange sdk.Dec `json:"inflation_rate_change" yaml:"inflation_rate_change"` // Deprecated: maximum annual change in inflation rate
	InflationMax        sdk.Dec `json:"inflation_max" yaml:"inflation_max"`                 // Deprecated: maximum inflation rate
	InflationMin        sdk.Dec `json:"inflation_min" yaml:"inflation_min"`                 // Deprecated: minimum inflation rate
	GoalBonded          sdk.Dec `json:"goal_bonded" yaml:"goal_bonded"`                     // Deprecated: goal of percent bonded atoms
	BlocksPerYear       uint64  `json:"blocks_per_year" yaml:"blocks_per_year"`             // blocks per year according to one block per 3s

	DeflationRate  sdk.Dec `json:"deflation_rate" yaml:"deflation_rate"` // deflation rate every DeflationEpoch
	DeflationEpoch uint64  `json:"deflation_epoch" yaml:"deflation_epoch"` // block number to deflate
	FarmProportion sdk.Dec `json:"farm_proportion" yaml:"farm_proportion"` // proportion of minted for farm
}

// ParamTable for minting module.
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	mintDenom string, inflationRateChange, inflationMax, inflationMin, goalBonded sdk.Dec, blocksPerYear uint64,
	deflationEpoch uint64, deflationRateChange, farmPropotion sdk.Dec,
) Params {

	return Params{
		MintDenom:      mintDenom,
		BlocksPerYear:  blocksPerYear,
		DeflationRate:  deflationRateChange,
		DeflationEpoch: deflationEpoch,
		FarmProportion: farmPropotion,
	}
}

// default minting module parameters
func DefaultParams() Params {
	return Params{
		MintDenom: sdk.DefaultBondDenom,
		//InflationRateChange: sdk.NewDecWithPrec(13, 2),
		//InflationMax:        sdk.NewDecWithPrec(20, 2),
		//InflationMin:        sdk.NewDecWithPrec(7, 2),
		//GoalBonded:          sdk.NewDecWithPrec(67, 2),
		BlocksPerYear:  uint64(60 * 60 * 8766 / 3), // assuming 3 second block times
		DeflationRate:  sdk.NewDecWithPrec(5, 1),
		DeflationEpoch: 3,                        // 3 years
		FarmProportion: sdk.NewDecWithPrec(5, 1), // 0.5
	}
}

// validate params
func (p Params) Validate() error {
	if err := validateMintDenom(p.MintDenom); err != nil {
		return err
	}
	if err := validateDeflationRate(p.DeflationRate); err != nil {
		return err
	}
	if err := validateDeflationEpoch(p.DeflationEpoch); err != nil {
		return err
	}
	if err := validateFarmProportion(p.FarmProportion); err != nil {
		return err
	}
	if err := validateBlocksPerYear(p.BlocksPerYear); err != nil {
		return err
	}

	return nil
}

func (p Params) String() string {
	return fmt.Sprintf(`Minting Params:
  Mint Denom:                     %s
  Deflation Rate Every %d Years:  %s
  Blocks Per Year:                %d
  Farm Proportion:                %s
`,
		p.MintDenom, p.DeflationEpoch, p.DeflationRate, p.BlocksPerYear, p.FarmProportion,
	)
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyMintDenom, &p.MintDenom, validateMintDenom),
		params.NewParamSetPair(KeyBlocksPerYear, &p.BlocksPerYear, validateBlocksPerYear),
		params.NewParamSetPair(KeyDeflationRate, &p.DeflationRate, validateDeflationRate),
		params.NewParamSetPair(KeyDeflationEpoch, &p.DeflationEpoch, validateDeflationEpoch),
		params.NewParamSetPair(KeyFarmProportion, &p.FarmProportion, validateFarmProportion),
	}
}

func validateMintDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return errors.New("mint denom cannot be blank")
	}
	if err := sdk.ValidateDenom(v); err != nil {
		return err
	}

	return nil
}

func validateInflationRateChange(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("inflation rate change cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("inflation rate change too large: %s", v)
	}

	return nil
}

func validateInflationMax(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("max inflation cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("max inflation too large: %s", v)
	}

	return nil
}

func validateInflationMin(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("min inflation cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("min inflation too large: %s", v)
	}

	return nil
}

func validateGoalBonded(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("goal bonded cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("goal bonded too large: %s", v)
	}

	return nil
}

func validateBlocksPerYear(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("blocks per year must be positive: %d", v)
	}

	return nil
}

func validateFarmProportion(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("Farm Proportion be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("Farm Proportion too large: %s", v)
	}

	return nil
}

func validateDeflationRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("Deflation Rate be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("Deflation Rate too large: %s", v)
	}

	return nil
}

func validateDeflationEpoch(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("Deflation Epoch must be positive: %d", v)
	}

	return nil
}
