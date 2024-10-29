//go:build !libsecp256k1_sdk
// +build !libsecp256k1_sdk

package secp256k1

import (
	"testing"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/stretchr/testify/require"
)

// Ensure that signature verification works, and that
// non-canonical signatures fail.
// Note: run with CGO_ENABLED=0 or go test -tags !cgo.
func TestSignatureVerificationAndRejectUpperS(t *testing.T) {
	msg := []byte("We have lingered long enough on the shores of the cosmic ocean.")
	for i := 0; i < 500; i++ {
		priv := GenPrivKey()
		sigStr, err := priv.Sign(msg)
		require.NoError(t, err)
		var r secp256k1.ModNScalar
		r.SetByteSlice(sigStr[:32])
		var s secp256k1.ModNScalar
		s.SetByteSlice(sigStr[32:64])
		require.False(t, s.IsOverHalfOrder())

		pub := priv.PubKey()
		require.True(t, pub.VerifySignature(msg, sigStr))

		// malleate:
		var S256 secp256k1.ModNScalar
		S256.SetByteSlice(secp256k1.S256().N.Bytes())
		s.Negate().Add(&S256)
		require.True(t, s.IsOverHalfOrder())
	}
}
