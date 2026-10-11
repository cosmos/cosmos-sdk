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
	// The 133/0/0, 60/1/7, and max-index seeds also match pqseed's scoped vectors:
	// https://github.com/iqlusioninc/crates/blob/9f43284850d205f6c501863603670b2ca1ffbddf/pqseed/tests/scoped.rs
	mnemonic := strings.Repeat("abandon ", 23) + "art"
	for _, tc := range []struct {
		passphrase, path, seed, pubHash string
	}{
		{"", "", "2a278082cc03c4d627e42b33008dab5ad957decf4422957766b55e2cffac9661", "91834016126aaf90cc683bab0f24dec823c19bf96580787b24dc2f66110a6e90"},
		{"TREZOR", "m/83696968'/505153'/118'/3'/7'/9'", "7f6d76f9c904d2df7cdc2f919911ccc7b98ba28c5cc1f8e0fd59ed889f35fbab", "2d29e6c1b900d9b9c70b9484424c3d4a8947c90545912583b3234f2f0b7fffdf"},
		{"café", "", "dd9ecfd95d9056216389a4227cb1ac040861ba76b350871acfc3c623ef317049", "3cad8f7b23a39aed0d9792746b60632eccf0cecafec525e4fc89dcdacb4546c8"},
		{"cafe\u0301", "", "dd9ecfd95d9056216389a4227cb1ac040861ba76b350871acfc3c623ef317049", "3cad8f7b23a39aed0d9792746b60632eccf0cecafec525e4fc89dcdacb4546c8"},
		{"", "m/83696968'/505153'/2147483647'/3'/2147483647'/2147483647'", "f13b209d10a691d8eb09e092d349ba3bc1cee7231a5030aa57e37af3a0beb999", "6857adae85762d4b14329cf549bf6882a9aac1a94e92725406e7c8943e5878d6"},
		{"", "m/83696968'/505153'/133'/3'/0'/0'", "14b0f8eaf217e75f06098c7e4a9655bde6d1af5b99293f5f1ee0bc270afb436e", "f6caf0b6afb7fe7afe337378b4c2864a75a4970568572f96b9db0b29d57bd85c"},
		{"", "m/83696968'/505153'/60'/3'/1'/7'", "b89d7e462b757aa53c2b4c19521a4c69ecceee8eed5905c4dad2f22dd6e2a95d", "b5b626b6856ff3e2a314008102c5457bdb18cb64a3c6110bf13acd13499a862d"},
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
	path, err := hd.CreateMlDsa65BIP85Path(118, 0, 0)
	require.NoError(t, err)
	seed, err := hd.MlDsa65BIP85.Derive()(mnemonic, "", path)
	require.NoError(t, err)
	require.Equal(t, "2a278082cc03c4d627e42b33008dab5ad957decf4422957766b55e2cffac9661", hex.EncodeToString(seed))
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
	for _, path := range []string{
		hdPath0, "m/83696968'/503131'/3'", "m/83696968'/118'/65'/0'/0'",
		"m/83696968'/505153'/118'/0'/0'/0'", "m/83696968'/505153'/118'/65'/0'/0'",
		"m/83696968'/505153'/118/3'/0'/0'", "m/83696968'/505153'/118'/3'/0'/0",
		"m/83696968'/505153'/118'/3'/0'/0'/1'", "m/83696968'/505153'/118'/3'/0'",
		"m/83696968'/505153'/2147483648'/3'/0'/0'",
		"m/83696968'/505153'/118'/3'/2147483648'/0'",
		"m/83696968'/505153'/118'/3'/0'/2147483648'",
	} {
		_, err := hd.MlDsa65BIP85.Derive()(mnemonic, "", path)
		require.Error(t, err)
	}
	for _, scope := range [][3]uint32{{1 << 31, 0, 0}, {118, 1 << 31, 0}, {118, 0, 1 << 31}} {
		_, err := hd.CreateMlDsa65BIP85Path(scope[0], scope[1], scope[2])
		require.Error(t, err)
	}
}
