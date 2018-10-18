package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// GenesisState - all distribution state that must be provided at genesis
type GenesisState struct {
	Minter Minter `json:"Minter"` // minter object
	Params Params `json:"params"` // inflation params
}

func NewGenesisState(minter Minter, params Params) GenesisState {
	return GenesisState{
		Minter: minter,
		Params: params,
	}
}

// get raw genesis raw message for testing
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Minter: InitialMinter(),
		Params: DefaultParams(),
	}
}

func InitGenesis(ctx sdk.Context, keeper Keeper, data types.GenesisState) {
	keeper.SetMinter(ctx, data.Minter)
	keeper.SetParams(ctx, data.Params)
}

// WriteGenesis returns a GenesisState for a given context and keeper. The
// GenesisState will contain the pool, and validator/delegator distribution info's
func WriteGenesis(ctx sdk.Context, keeper Keeper) types.GenesisState {

	minter := keeper.GetMinter(ctx, data.Minter)
	params := keeper.GetParams(ctx, data.Params)
	return NewGenesisState(mintDenom, minter, params)
}
