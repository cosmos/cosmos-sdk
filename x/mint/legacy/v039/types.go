package v039

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	ModuleName = "mint"
)

type (
	// Minter represents the minting state.
	Minter struct {
		Inflation        sdk.Dec `json:"inflation" yaml:"inflation"`                 // current annual inflation rate
		AnnualProvisions sdk.Dec `json:"annual_provisions" yaml:"annual_provisions"` // current annual expected provisions
	}

	// mint parameters
	Params struct {
		MintDenom            string  `json:"mint_denom" yaml:"mint_denom"`                           // type of coin to mint
		InflationRateChange  sdk.Dec `json:"inflation_rate_change" yaml:"inflation_rate_change"`     // maximum annual change in inflation rate
		InflationMax         sdk.Dec `json:"inflation_max" yaml:"inflation_max"`                     // maximum inflation rate
		InflationMin         sdk.Dec `json:"inflation_min" yaml:"inflation_min"`                     // minimum inflation rate
		GoalBonded           sdk.Dec `json:"goal_bonded" yaml:"goal_bonded"`                         // goal of percent bonded atoms
		BlocksPerYear        uint64  `json:"blocks_per_year" yaml:"blocks_per_year"`                 // expected blocks per year
		MaxMintableAmount    uint64  `json:"max_mintable_amount" yaml:"max_mintable_amount"`         // expected max_mintable_amount
		MintedAmountPerBlock sdk.Dec `json:"minted_amount_per_block" yaml:"minted_amount_per_block"` // expected minted_amount_per_block
		YearlyReduction      sdk.Dec `json:"yearly_reduction" yaml:"yearly_reduction"`               // expected yearly_reduction
	}

	// GenesisState - minter state
	GenesisState struct {
		Minter Minter `json:"minter" yaml:"minter"` // minter object
		Params Params `json:"params" yaml:"params"` // inflation params
	}
)
