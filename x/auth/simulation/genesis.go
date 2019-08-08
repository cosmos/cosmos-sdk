package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/simulation"
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
func GenMaxMemoChars(cdc *codec.Codec, r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 100, 200))
}

// GenTxSigLimit randomized TxSigLimit
func GenTxSigLimit(cdc *codec.Codec, r *rand.Rand) uint64 {
	return uint64(r.Intn(7) + 1)
}

// GenTxSizeCostPerByte randomized TxSizeCostPerByte
func GenTxSizeCostPerByte(cdc *codec.Codec, r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 5, 15))
}

// GenSigVerifyCostED25519 randomized SigVerifyCostED25519
func GenSigVerifyCostED25519(cdc *codec.Codec, r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 500, 1000))
}

// GenSigVerifyCostSECP256K1 randomized SigVerifyCostSECP256K1
func GenSigVerifyCostSECP256K1(cdc *codec.Codec, r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 500, 1000))
}

// GenAuthGenesisState generates a random GenesisState for auth
func GenAuthGenesisState(cdc *codec.Codec, r *rand.Rand, genesisState map[string]json.RawMessage) {

	maxMemoChars := GenMaxMemoChars(cdc, r)
	txSigLimit := GenTxSigLimit(cdc, r)
	txSizeCostPerByte := GenTxSizeCostPerByte(cdc, r)
	sigVerifyCostED25519 := GenSigVerifyCostED25519(cdc, r)
	sigVerifyCostSECP256K1 := GenSigVerifyCostSECP256K1(cdc, r)

	authGenesis := auth.NewGenesisState(
		auth.NewParams(maxMemoChars, txSigLimit, txSizeCostPerByte,
			sigVerifyCostED25519, sigVerifyCostSECP256K1),
	)

	fmt.Printf("Selected randomly generated auth parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, authGenesis.Params))
	genesisState[auth.ModuleName] = cdc.MustMarshalJSON(authGenesis)
}
