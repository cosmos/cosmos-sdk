package ante

import (
	"encoding/hex"

	"github.com/tendermint/tendermint/crypto/secp256k1"
)

var (
	// simulation signature values used to estimate gas consumption
	simSecp256k1Pubkey secp256k1.PubKeySecp256k1
	simSecp256k1Sig    [64]byte
)

func init() {
	// This decodes a valid hex string into a sepc256k1Pubkey for use in transaction simulation
	bz, _ := hex.DecodeString("035AD6810A47F073553FF30D2FCC7E0D3B1C0B74B61A1AAA2582344037151E143A")
	copy(simSecp256k1Pubkey[:], bz)
}
