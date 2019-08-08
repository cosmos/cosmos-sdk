package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// GenSupplyGenesisState generates a random GenesisState for supply
func GenSupplyGenesisState(cdc *codec.Codec, _ *rand.Rand, _, genesisState map[string]json.RawMessage,
	amount, numInitiallyBonded, numAccs int64) {
	totalSupply := sdk.NewInt(amount * (numAccs + numInitiallyBonded))
	supplyGenesis := supply.NewGenesisState(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, totalSupply)))

	fmt.Printf("Generated supply parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, supplyGenesis))
	genesisState[supply.ModuleName] = cdc.MustMarshalJSON(supplyGenesis)
}
