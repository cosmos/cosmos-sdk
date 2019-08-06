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

// GenMaxMemoChars randomized MaxMemoChars
func GenMaxMemoChars(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams) (maxMemoChars uint64) {
	ap.GetOrGenerate(cdc, MaxMemoChars, &maxMemoChars, r,
		func(r *rand.Rand) {
			maxMemoChars = uint64(simutil.RandIntBetween(r, 100, 200))
		})
	return
}

// GenTxSigLimit randomized TxSigLimit
func GenTxSigLimit(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams) (txSigLimit uint64) {
	ap.GetOrGenerate(cdc, TxSigLimit, &txSigLimit, r,
		func(r *rand.Rand) {
			txSigLimit = uint64(r.Intn(7) + 1)
		})
	return
}

// GenTxSizeCostPerByte randomized TxSizeCostPerByte
func GenTxSizeCostPerByte(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams) (txSizeCostPerByte uint64) {
	ap.GetOrGenerate(cdc, TxSizeCostPerByte, &txSizeCostPerByte, r,
		func(r *rand.Rand) {
			txSizeCostPerByte = uint64(simutil.RandIntBetween(r, 5, 15))
		})
	return
}

// GenSigVerifyCostED25519 randomized SigVerifyCostED25519
func GenSigVerifyCostED25519(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams) (sigVerifyCostED25519 uint64) {
	ap.GetOrGenerate(cdc, SigVerifyCostED25519, &sigVerifyCostED25519, r,
		func(r *rand.Rand) {
			sigVerifyCostED25519 = uint64(simutil.RandIntBetween(r, 500, 1000))
		})
	return
}

// GenSigVerifyCostSECP256K1 randomized SigVerifyCostSECP256K1
func GenSigVerifyCostSECP256K1(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams) (sigVerifyCostSECP256K1 uint64) {
	ap.GetOrGenerate(cdc, SigVerifyCostSECP256K1, &sigVerifyCostSECP256K1, r,
		func(r *rand.Rand) {
			sigVerifyCostSECP256K1 = uint64(simutil.RandIntBetween(r, 500, 1000))
		})
	return
}

// GenAuthGenesisState generates a random GenesisState for auth
func GenAuthGenesisState(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {

	maxMemoChars := GenMaxMemoChars(cdc, r, ap)
	txSigLimit := GenTxSigLimit(cdc, r, ap)
	txSizeCostPerByte := GenTxSizeCostPerByte(cdc, r, ap)
	sigVerifyCostED25519 := GenSigVerifyCostED25519(cdc, r, ap)
	sigVerifyCostSECP256K1 := GenSigVerifyCostSECP256K1(cdc, r, ap)

	authGenesis := auth.NewGenesisState(
		auth.NewParams(maxMemoChars, txSigLimit, txSizeCostPerByte,
			sigVerifyCostED25519, sigVerifyCostSECP256K1),
	)

	fmt.Printf("Selected randomly generated auth parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, authGenesis.Params))
	genesisState[auth.ModuleName] = cdc.MustMarshalJSON(authGenesis)
}
