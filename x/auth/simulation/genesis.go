package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
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

// RandomizedGenState generates a random GenesisState for auth
func RandomizedGenState(input *module.GeneratorInput) {

	maxMemoChars := GenMaxMemoChars(input.Cdc, input.R)
	txSigLimit := GenTxSigLimit(input.Cdc, input.R)
	txSizeCostPerByte := GenTxSizeCostPerByte(input.Cdc, input.R)
	sigVerifyCostED25519 := GenSigVerifyCostED25519(input.Cdc, input.R)
	sigVerifyCostSECP256K1 := GenSigVerifyCostSECP256K1(input.Cdc, input.R)

	authGenesis := types.NewGenesisState(
		types.NewParams(maxMemoChars, txSigLimit, txSizeCostPerByte,
			sigVerifyCostED25519, sigVerifyCostSECP256K1),
	)

	fmt.Printf("Selected randomly generated auth parameters:\n%s\n", codec.MustMarshalJSONIndent(input.Cdc, authGenesis.Params))
	input.GenState[types.ModuleName] = input.Cdc.MustMarshalJSON(authGenesis)
}
