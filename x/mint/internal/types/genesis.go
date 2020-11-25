package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState - minter state
type GenesisState struct {
	Minter MinterCustom `json:"minter" yaml:"minter"` // minter object
	Params Params `json:"params" yaml:"params"` // inflation params

	OriginalMintedPerBlock sdk.Dec      `json:"original_minted_per_block" yaml:"original_minted_per_block"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(minter MinterCustom, params Params, originalMintedPerBlock sdk.Dec) GenesisState {
	return GenesisState{
		Minter: minter,
		Params: params,

		OriginalMintedPerBlock: originalMintedPerBlock,
	}
}

func DefaultOriginalMintedPerBlock() sdk.Dec {
	return sdk.MustNewDecFromStr("1")
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Minter: DefaultInitialMinterCustom(),
		Params: DefaultParams(),
		OriginalMintedPerBlock: DefaultOriginalMintedPerBlock(),
	}
}

// ValidateGenesis validates the provided genesis state to ensure the
// expected invariants holds.
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}

	return ValidateMinterCustom(data.Minter)
}
