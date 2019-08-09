package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/bank/internal/types"
)

// Simulation parameter constants
const (
	SendEnabled = "send_enabled"
)

// GenSendEnabled randomized SendEnabled
func GenSendEnabled(cdc *codec.Codec, r *rand.Rand) bool {
	return r.Int63n(2) == 0
}

// GenBankGenesisState generates a random GenesisState for bank
func GenBankGenesisState(cdc *codec.Codec, r *rand.Rand, genesisState map[string]json.RawMessage) {
	sendEnabled := GenSendEnabled(cdc, r)

	bankGenesis := types.NewGenesisState(sendEnabled)

	fmt.Printf("Selected randomly generated bank parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, bankGenesis))
	genesisState[types.ModuleName] = cdc.MustMarshalJSON(bankGenesis)
}
