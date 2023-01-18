package keys

import (
	"math/big"

	"github.com/cosmos/cosmos-sdk/crypto/keys/internal/ecdsa"
)

// Replicates https://github.com/cosmos/cosmos-sdk/blob/44fbb0df9cea049d588e76bf930177d777552cf3/crypto/ledger/ledger_secp256k1.go#L228
// DO NOT USE. This is a temporary workaround that is cleaned-up in v0.47+
func IsOverHalfOrder(sigS *big.Int) bool {
	return !ecdsa.IsSNormalized(sigS)
}
