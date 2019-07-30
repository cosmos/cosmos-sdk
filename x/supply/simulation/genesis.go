package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// GenSupplyGenesisState generates a random GenesisState for supply
func GenSupplyGenesisState(cdc *codec.Codec, amount, numInitiallyBonded, numAccs int64, genesisState map[string]json.RawMessage) {
	totalSupply := sdk.NewInt(amount * (numAccs + numInitiallyBonded))
	supplyGenesis := supply.NewGenesisState(
		supply.NewSupply(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, totalSupply))),
	)

	fmt.Printf("Generated supply parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, supplyGenesis))
	genesisState[supply.ModuleName] = cdc.MustMarshalJSON(supplyGenesis)
}
