package simulation

// DONTCOVER

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/supply/internal/types"
)

// RandomizedGenState generates a random GenesisState for supply
func RandomizedGenState(input *module.GeneratorInput) {

	numAccs := int64(len(input.Accounts))
	totalSupply := sdk.NewInt(input.InitialStake * (numAccs + input.NumBonded))
	supplyGenesis := types.NewGenesisState(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, totalSupply)))

	fmt.Printf("Generated supply parameters:\n%s\n", codec.MustMarshalJSONIndent(input.Cdc, supplyGenesis))
	input.GenState[types.ModuleName] = input.Cdc.MustMarshalJSON(supplyGenesis)
}
