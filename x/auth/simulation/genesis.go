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
func GenMaxMemoChars(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 100, 200))
}

// GenTxSigLimit randomized TxSigLimit
func GenTxSigLimit(r *rand.Rand) uint64 {
	return uint64(r.Intn(7) + 1)
}

// GenTxSizeCostPerByte randomized TxSizeCostPerByte
func GenTxSizeCostPerByte(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 5, 15))
}

// GenSigVerifyCostED25519 randomized SigVerifyCostED25519
func GenSigVerifyCostED25519(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 500, 1000))
}

// GenSigVerifyCostSECP256K1 randomized SigVerifyCostSECP256K1
func GenSigVerifyCostSECP256K1(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 500, 1000))
}

// RandomizedGenState generates a random GenesisState for auth
func RandomizedGenState(input *module.GeneratorInput) {

	var (
		maxMemoChars           uint64
		txSigLimit             uint64
		txSizeCostPerByte      uint64
		sigVerifyCostED25519   uint64
		sigVerifyCostSECP256K1 uint64
	)

	input.AppParams.GetOrGenerate(input.Cdc, MaxMemoChars, &maxMemoChars, input.R,
		func(r *rand.Rand) { maxMemoChars = GenMaxMemoChars(input.R) })

	input.AppParams.GetOrGenerate(input.Cdc, TxSigLimit, &txSigLimit, input.R,
		func(r *rand.Rand) { txSigLimit = GenTxSigLimit(input.R) })

	input.AppParams.GetOrGenerate(input.Cdc, TxSizeCostPerByte, &txSizeCostPerByte, input.R,
		func(r *rand.Rand) { txSizeCostPerByte = GenTxSizeCostPerByte(input.R) })

	input.AppParams.GetOrGenerate(input.Cdc, SigVerifyCostED25519, &sigVerifyCostED25519, input.R,
		func(r *rand.Rand) { sigVerifyCostED25519 = GenSigVerifyCostED25519(input.R) })

	input.AppParams.GetOrGenerate(input.Cdc, SigVerifyCostSECP256K1, &sigVerifyCostSECP256K1, input.R,
		func(r *rand.Rand) { sigVerifyCostED25519 = GenSigVerifyCostSECP256K1(input.R) })

	authGenesis := types.NewGenesisState(
		types.NewParams(maxMemoChars, txSigLimit, txSizeCostPerByte,
			sigVerifyCostED25519, sigVerifyCostSECP256K1),
	)

	fmt.Printf("Selected randomly generated auth parameters:\n%s\n", codec.MustMarshalJSONIndent(input.Cdc, authGenesis.Params))
	input.GenState[types.ModuleName] = input.Cdc.MustMarshalJSON(authGenesis)
}
