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
func RandomizedGenState(simState *module.SimulationState) {
	var maxMemoChars uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MaxMemoChars, &maxMemoChars, simState.Rand,
		func(r *rand.Rand) { maxMemoChars = GenMaxMemoChars(r) },
	)

	var txSigLimit uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, TxSigLimit, &txSigLimit, simState.Rand,
		func(r *rand.Rand) { txSigLimit = GenTxSigLimit(r) },
	)

	var txSizeCostPerByte uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, TxSizeCostPerByte, &txSizeCostPerByte, simState.Rand,
		func(r *rand.Rand) { txSizeCostPerByte = GenTxSizeCostPerByte(r) },
	)

	var sigVerifyCostED25519 uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, SigVerifyCostED25519, &sigVerifyCostED25519, simState.Rand,
		func(r *rand.Rand) { sigVerifyCostED25519 = GenSigVerifyCostED25519(r) },
	)

	var sigVerifyCostSECP256K1 uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, SigVerifyCostSECP256K1, &sigVerifyCostSECP256K1, simState.Rand,
		func(r *rand.Rand) { sigVerifyCostED25519 = GenSigVerifyCostSECP256K1(r) },
	)

	authGenesis := types.NewGenesisState(
		types.NewParams(maxMemoChars, txSigLimit, txSizeCostPerByte,
			sigVerifyCostED25519, sigVerifyCostSECP256K1),
	)

	fmt.Printf("Selected randomly generated auth parameters:\n%s\n", codec.MustMarshalJSONIndent(simState.Cdc, authGenesis.Params))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(authGenesis)
}
