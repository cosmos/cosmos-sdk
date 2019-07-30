package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// GenBankGenesisState generates a random GenesisState for bank
func GenBankGenesisState(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	bankGenesis := bank.NewGenesisState(
		func(r *rand.Rand) bool {
			var v bool
			ap.GetOrGenerate(cdc, simulation.SendEnabled, &v, r,
				func(r *rand.Rand) {
					v = simulation.ModuleParamSimulator[simulation.SendEnabled](r).(bool)
				})
			return v
		}(r),
	)

	fmt.Printf("Selected randomly generated bank parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, bankGenesis))
	genesisState[bank.ModuleName] = cdc.MustMarshalJSON(bankGenesis)
}
