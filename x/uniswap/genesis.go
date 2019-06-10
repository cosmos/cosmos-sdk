package uniswap

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState - uniswap genesis state
type GenesisState struct {
	NativeAssetDenom string    `json:"native_asset_denom"`
	FeeParams        FeeParams `json:"fee_params"`
}

// NewGenesisState is the constructor function for GenesisState
func NewGenesisState(nativeAssetDenom string, feeParams FeeParams) GenesisState {
	return GenesisState{
		NativeAssetDenom: nativeAssetDenom,
		FeeParams:        feeParams,
	}
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() GenesisState {
	return NewGenesisState(sdk.DefaultBondDenom, DefaultParams())
}

// InitGenesis new uniswap genesis
// TODO
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {

}

// ExportGenesis returns a GenesisState for a given context and keeper.
// TODO
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	return NewGenesisState(sdk.DefaultBondDenom)
}

// ValidateGenesis - placeholder function
func ValidateGenesis(data GenesisState) error {
	if strings.TrimSpace(data.NativeAssetDenom) == "" {
		fmt.Errorf("no native asset denomination provided")
	}
	if err := data.FeeParams.Validate(); err != nil {
		return err
	}
	return nil
}
