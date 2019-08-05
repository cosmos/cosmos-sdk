package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simutil "github.com/cosmos/cosmos-sdk/x/simulation/util"
)

// Simulation parameter constants
const (
	MaxMemoChars           = "max_memo_characters"
	TxSigLimit             = "tx_sig_limit"
	TxSizeCostPerByte      = "tx_size_cost_per_byte"
	SigVerifyCostED25519   = "sig_verify_cost_ed25519"
	SigVerifyCostSECP256K1 = "sig_verify_cost_secp256k1"
)

// GenAuthGenesisState generates a random GenesisState for auth
func GenAuthGenesisState(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	
	var maxMemoChars uint64
	ap.GetOrGenerate(cdc, MaxMemoChars, &maxMemoChars, r,
		func(r *rand.Rand) {
			maxMemoChars = uint64(simutil.RandIntBetween(r, 100, 200))
		})

		var txSigLimit uint64
		ap.GetOrGenerate(cdc, TxSigLimit, &txSigLimit, r,
			func(r *rand.Rand) {
				txSigLimit = uint64(r.Intn(7) + 1)
			})
	
		var txSizeCostPerByte uint64
		ap.GetOrGenerate(cdc, TxSizeCostPerByte, &txSizeCostPerByte, r,
			func(r *rand.Rand) {
				txSizeCostPerByte = uint64(simutil.RandIntBetween(r, 5, 15))
			})

		var sigVerifyCostED25519 uint64
		ap.GetOrGenerate(cdc, SigVerifyCostED25519, &sigVerifyCostED25519, r,
			func(r *rand.Rand) {
				sigVerifyCostED25519 = uint64(simutil.RandIntBetween(r, 500, 1000))
			})
		
			var sigVerifyCostSECP256K1 uint64
			ap.GetOrGenerate(cdc, SigVerifyCostSECP256K1, &sigVerifyCostSECP256K1, r,
				func(r *rand.Rand) {
					sigVerifyCostSECP256K1 = uint64(simutil.RandIntBetween(r, 500, 1000))
				})
	
	authGenesis := auth.NewGenesisState(
		auth.NewParams(maxMemoChars, txSigLimit, txSizeCostPerByte,
			sigVerifyCostED25519, sigVerifyCostSECP256K1),
	)

	fmt.Printf("Selected randomly generated auth parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, authGenesis.Params))
	genesisState[auth.ModuleName] = cdc.MustMarshalJSON(authGenesis)
}
