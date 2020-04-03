package privkey

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaults(t *testing.T) {
	require.Equal(t, PubKeyType("multi"), MultiType)
	require.Equal(t, PubKeyType("secp256k1"), Secp256k1Type)
	require.Equal(t, PubKeyType("ed25519"), Ed25519Type)
	require.Equal(t, PubKeyType("sr25519"), Sr25519Type)
}
