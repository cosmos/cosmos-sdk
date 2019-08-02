package simulation

import "math/rand"

// Simulation parameter constants
const (
	MaxMemoChars           = "max_memo_characters"
	TxSigLimit             = "tx_sig_limit"
	TxSizeCostPerByte      = "tx_size_cost_per_byte"
	SigVerifyCostED25519   = "sig_verify_cost_ed25519"
	SigVerifyCostSECP256K1 = "sig_verify_cost_secp256k1"
)

// GenParams generates random auth parameters
func GenParams(paramSims map[string]func(r *rand.Rand) interface{}) {
	paramSims[MaxMemoChars] = func(r *rand.Rand) interface{} {
		return uint64(RandIntBetween(r, 100, 200))
	}

	paramSims[TxSigLimit] = func(r *rand.Rand) interface{} {
		return uint64(r.Intn(7) + 1)
	}

	paramSims[TxSizeCostPerByte] = func(r *rand.Rand) interface{} {
		return uint64(RandIntBetween(r, 5, 15))
	}

	paramSims[SigVerifyCostED25519] = func(r *rand.Rand) interface{} {
		return uint64(RandIntBetween(r, 500, 1000))
	}

	paramSims[SigVerifyCostSECP256K1] = func(r *rand.Rand) interface{} {
		return uint64(RandIntBetween(r, 500, 1000))
	}
}
