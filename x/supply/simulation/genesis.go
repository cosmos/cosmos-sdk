package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply/internal/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// RandomizedGenState generates a random GenesisState for supply
func RandomizedGenState(cdc *codec.Codec, _ *rand.Rand, genesisState map[string]json.RawMessage,
	accs []simulation.Account, amount, numInitiallyBonded int64) {
	
	numAccs := int64(len(accs))
	totalSupply := sdk.NewInt(amount * (numAccs + numInitiallyBonded))
	supplyGenesis := types.NewGenesisState(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, totalSupply)))

	fmt.Printf("Generated supply parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, supplyGenesis))
	genesisState[types.ModuleName] = cdc.MustMarshalJSON(supplyGenesis)
}
