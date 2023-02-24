package mintkey

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto"
	pdkdf2 "golang.org/x/crypto/pbkdf2"
)

func BenchmarkBcryptGenerateFromPassword(b *testing.B) {
	passphrase := []byte("passphrase")
	for securityParam := 9; securityParam < 16; securityParam++ {
		param := securityParam
		b.Run(fmt.Sprintf("benchmark-security-param-%d", param), func(b *testing.B) {
			saltBytes := crypto.CRandBytes(16)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := pdkdf2.Key([]byte(passphrase), saltBytes, param, 24, sha256.New)
				require.NotNil(b, key)
			}
		})
	}
}
