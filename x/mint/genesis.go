package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

// GenesisState - minter state
type GenesisState struct {
	Minter Minter       `json:"minter"` // minter object
	Params types.Params `json:"params"` // inflation params
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(minter Minter, params Params) GenesisState {
	return GenesisState{
		Minter: minter,
		Params: params,
	}
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Minter: DefaultInitialMinter(),
		Params: DefaultParams(),
	}
}

// InitGenesis new mint genesis
func InitGenesis(ctx sdk.Context, keeper Keeper, ak AccountKeeper, data GenesisState) {
	// check if the module account exists and create it if not
	moduleAcc := keeper.GetMinterAccount(ctx)
	if moduleAcc == nil {
		moduleAcc = supply.NewModuleMinterAccount(ModuleName)
		if err := moduleAcc.SetAccountNumber(ak.GetNextAccountNumber(ctx)); err != nil {
			panic(err)
		}
		keeper.SetMinterAccount(ctx, moduleAcc)
	}

	keeper.SetMinter(ctx, data.Minter)
	keeper.SetParams(ctx, data.Params)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {

	minter := keeper.GetMinter(ctx)
	params := keeper.GetParams(ctx)
	return NewGenesisState(minter, params)
}

// ValidateGenesis validates the provided genesis state to ensure the
// expected invariants holds.
func ValidateGenesis(data GenesisState) error {
	err := ValidateParams(data.Params)
	if err != nil {
		return err
	}
	err = ValidateMinter(data.Minter)
	if err != nil {
		return err
	}
	return nil
}
