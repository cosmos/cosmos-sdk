package eth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/big"

	secp "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1/internal/secp256k1"
)

// ----------------------------------------------------------------------------
// Helper Functions

func paddedBigBytes(bigint *big.Int, n int) []byte {
	if bigint.BitLen()/8 >= n {
		return bigint.Bytes()
	}
	ret := make([]byte, n)
	i := len(ret)
	for _, d := range bigint.Bits() {
		for j := 0; j < wordBytes && i > 0; j++ {
			i--
			ret[i] = byte(d)
			d >>= 8
		}
	}
	return ret
}

func FromECDSA(priv *ecdsa.PrivateKey) []byte {
	if priv == nil {
		return nil
	}
	return paddedBigBytes(priv.D, priv.Params().BitSize/8)
}

func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(secp.S256(), pub.X, pub.Y)
}
