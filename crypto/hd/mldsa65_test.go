package hd_test

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mldsa65"
)

const (
	testMnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	hdPath0      = "m/44'/118'/0'/0/0"
	hdPath1      = "m/44'/118'/0'/0/1"
)

func TestMlDsa65Name(t *testing.T) {
	require.Equal(t, hd.MlDsa65Type, hd.MlDsa65.Name())
}

func TestMlDsa65DeriveAndGenerate(t *testing.T) {
	derive := hd.MlDsa65.Derive()
	gen := hd.MlDsa65.Generate()

	seed, err := derive(testMnemonic, "", hdPath0)
	require.NoError(t, err)
	require.Len(t, seed, 32) // matches mldsa65 SeedSize
	// Freeze legacy recovery independently of the BIP85 implementation.
	require.Equal(t, "c4a48e2fce1481cd3294b4490f6678090ea98d3d0e5cd984558ab0968741b104", hex.EncodeToString(seed))

	priv := gen(seed)
	require.NotNil(t, priv)
	_, ok := priv.(*mldsa65.PrivKey)
	require.True(t, ok)
}

func TestMlDsa65GenerateZeroSeedNoPanic(t *testing.T) {
	require.NotPanics(t, func() {
		priv := hd.MlDsa65.Generate()(make([]byte, 32))
		require.NotNil(t, priv)
		require.Equal(t, "ml_dsa_65", priv.PubKey().Type())
	})
}

func TestMlDsa65DerivationDeterministic(t *testing.T) {
	derive := hd.MlDsa65.Derive()
	gen := hd.MlDsa65.Generate()

	mk := func(path string) []byte {
		seed, err := derive(testMnemonic, "", path)
		require.NoError(t, err)
		return gen(seed).PubKey().Bytes()
	}

	// Same mnemonic+path -> same key (recovery).
	require.Equal(t, mk(hdPath0), mk(hdPath0))
	// Different path -> different account (path salt).
	require.NotEqual(t, mk(hdPath0), mk(hdPath1))
}

func TestMlDsa65BIP85Vectors(t *testing.T) {
	// End-to-end vectors generated independently with @scure/bip39,
	// @scure/bip32, Node's HMAC-SHA512, and @noble/post-quantum/ml-dsa.
	mnemonic := strings.Repeat("abandon ", 23) + "art"
	for _, tc := range []struct {
		passphrase, path, seed, pubHash string
	}{
		{"", "", "b923acb133a6335d4ba2fe0a9c6f5ad7e8cc9b1c18bab85f4e33d3dde7a24ef5", "52ba137931c9027dab4ae828ecb89d346d5b8330c4901b3087a885cf3b60b83a"},
		{"TREZOR", "m/83696968'/118'/65'/7'/9'", "a0725f164ac29971de71b079f8a331c3f9d0fdec93d14e697d79978c0cf7cdc5", "a7249027b2956307c88c22827f280ce9c268ab386b1261b2d8dccc30b18172e7"},
		{"café", "", "82086af0688409fa5f1b5aa7849323e3973f797e557f2ee5ec67ca3805307ad9", "e4d484144a10b4eda7800ecd52d9108a43de8fe908537f2b63cb69fc668748e9"},
		{"cafe\u0301", "", "82086af0688409fa5f1b5aa7849323e3973f797e557f2ee5ec67ca3805307ad9", "e4d484144a10b4eda7800ecd52d9108a43de8fe908537f2b63cb69fc668748e9"},
		{"", "m/83696968'/118'/65'/2147483647'/2147483647'", "2bf751da53983385b5d52b4d3fddd0ded724a647f43cf093d14a1d729e0939da", "328afa1ede1ab2604299712c2d1635fdacd8a63effc40127287358d771759dba"},
	} {
		t.Run(tc.passphrase+tc.path, func(t *testing.T) {
			seed, err := hd.MlDsa65BIP85.Derive()(mnemonic, tc.passphrase, tc.path)
			require.NoError(t, err)
			require.Equal(t, tc.seed, hex.EncodeToString(seed))
			priv := hd.MlDsa65BIP85.Generate()(seed)
			hash := sha256.Sum256(priv.PubKey().Bytes())
			require.Equal(t, tc.pubHash, hex.EncodeToString(hash[:]))
		})
	}
	path, err := hd.CreateMlDsa65BIP85Path(0, 0)
	require.NoError(t, err)
	seed, err := hd.MlDsa65BIP85.Derive()(mnemonic, "", path)
	require.NoError(t, err)
	require.Equal(t, "b923acb133a6335d4ba2fe0a9c6f5ad7e8cc9b1c18bab85f4e33d3dde7a24ef5", hex.EncodeToString(seed))
	legacy, err := hd.MlDsa65.Derive()(mnemonic, "", path)
	require.NoError(t, err)
	require.NotEqual(t, legacy, seed, "BIP85 must not expose the raw BIP32 child key")
}

func TestMlDsa65BIP85RejectsInvalidInput(t *testing.T) {
	mnemonic := strings.Repeat("abandon ", 23) + "art"
	for _, badMnemonic := range []string{"", testMnemonic, strings.Repeat("abandon ", 24), strings.Repeat("invalid ", 24)} {
		_, err := hd.MlDsa65BIP85.Derive()(badMnemonic, "", "")
		require.Error(t, err)
	}
	for _, path := range []string{hdPath0, "m/83696968'/503131'/0'", "m/83696968'/118'/44'/0'/0'", "m/83696968'/118'/65'/0'/0", "m/83696968'/118'/65'/0'/0'/1'"} {
		_, err := hd.MlDsa65BIP85.Derive()(mnemonic, "", path)
		require.Error(t, err)
	}
	for _, pair := range [][2]uint32{{1 << 31, 0}, {0, 1 << 31}} {
		_, err := hd.CreateMlDsa65BIP85Path(pair[0], pair[1])
		require.Error(t, err)
	}
}
