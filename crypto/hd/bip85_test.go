package hd_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/cosmos/go-bip39"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
)

func TestBIP85Vectors(t *testing.T) {
	// The master private scalar and chain code decoded from the xprv in
	// https://bips.dev/85/#test-vectors. Expected entropy is copied from BIP85.
	masterBytes, err := hex.DecodeString("3f15e5d852dc2e9ba5e9fe189a8dd2e1547badef5b563bbe6579fc6807d80ed9")
	require.NoError(t, err)
	chainBytes, err := hex.DecodeString("1b67969d1ec69bdfeeae43213da8460ba34b92d0788c8f7bfcfa44906e8a589c")
	require.NoError(t, err)
	for _, tc := range []struct{ path, entropy string }{
		{"m/83696968'/0'/0'", "efecfbccffea313214232d29e71563d941229afb4338c21f9517c41aaa0d16f00b83d2a09ef747e7a64e8e2bd5a14869e693da66ce94ac2da570ab7ee48618f7"},
		{"m/83696968'/0'/1'", "70c6e3e8ebee8dc4c0dbba66076819bb8c09672527c4277ca8729532ad711872218f826919f6b67218adde99018a6df9095ab2b58d803b5b93ec9802085a690e"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			for _, size := range []int{1, 32, 48, 64} {
				entropy, err := hd.DeriveBIP85Entropy([32]byte(masterBytes), [32]byte(chainBytes), tc.path, size)
				require.NoError(t, err)
				require.Equal(t, tc.entropy[:size*2], hex.EncodeToString(entropy))
			}
		})
	}
}

func TestBIP85ProjectElevenPaths(t *testing.T) {
	// Independent vectors from @scure/bip32 + @scure/bip39 and HMAC-SHA512,
	// using the article's exact application/algorithm IDs and seed lengths.
	seed := bip39.NewSeed(strings.Repeat("abandon ", 23)+"art", "")
	master, chain := hd.ComputeMastersFromSeed(seed)
	for _, tc := range []struct {
		path    string
		size    int
		entropy string
	}{
		{"m/83696968'/503131'/0'", 32, "e727e39395418f1af9ba73362ea790007d9f95512e8b28ace7f447240ccc313c"},
		{"m/83696968'/503131'/2'", 48, "7e76dfee07ba669b21ab0afc238896714463b45273b3d8ef8ebbba678814c454f51157f51515d2df8b111290b776d662"},
	} {
		entropy, err := hd.DeriveBIP85Entropy(master, chain, tc.path, tc.size)
		require.NoError(t, err)
		require.Equal(t, tc.entropy, hex.EncodeToString(entropy))
	}
}

func TestBIP85RejectsInvalidInput(t *testing.T) {
	master, chain := hd.ComputeMastersFromSeed([]byte("BIP85 test seed"))
	for _, path := range []string{
		"", "m", "m/83696968'", "m/83696968'/0'", "83696968'/0'/0'",
		"m/44'/0'/0'", "m/83696968/0'/0'", "m/83696968'/0/0'",
		"m/83696968'/0'/0", "m/83696968'/0'/0'/", "m/83696968'//0'",
		"m/83696968'/-1'/0'", "m/83696968'/2147483648'/0'",
		"m/83696968'/0'/2147483648'", "m/83696968'/0'/x'",
		"m/83696968'/0'/0''", "m/83696968'" + strings.Repeat("/0'", 255),
	} {
		t.Run(path, func(t *testing.T) {
			entropy, err := hd.DeriveBIP85Entropy(master, chain, path, 32)
			require.Error(t, err)
			require.Nil(t, entropy)
		})
	}
	for _, size := range []int{-1, 0, 65} {
		entropy, err := hd.DeriveBIP85Entropy(master, chain, "m/83696968'/0'/0'", size)
		require.Error(t, err)
		require.Nil(t, entropy)
	}
	// Reject zero and out-of-range master scalars, rather than reducing mod N.
	order, err := hex.DecodeString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141")
	require.NoError(t, err)
	for _, invalid := range [][32]byte{{}, [32]byte(order)} {
		entropy, err := hd.DeriveBIP85Entropy(invalid, chain, "m/83696968'/0'/0'", 32)
		require.Error(t, err)
		require.Nil(t, entropy)
	}
}
