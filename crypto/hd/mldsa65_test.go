package hd_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
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

func TestMlDsa65PQHDVectors(t *testing.T) {
	// Unmodified scheme-201 vectors from pqdh/pqhd main, with its MIT license
	// in testdata/pqhd-LICENSE-MIT. This pins the pre-v1 registry assignment:
	// https://github.com/pqdh/pqhd/blob/befdce14cecbb5d485d3239982bfa5b207cdb922/test-vectors/201.json
	data, err := os.ReadFile("testdata/pqhd-mldsa65.json")
	require.NoError(t, err)
	var vectors struct {
		Cases []struct {
			Mnemonic, Passphrase, Seed, Path, Output string
			CoinType                                 uint32 `json:"coin_type"`
			Account, Index                           uint32
			Entropy                                  string `json:"bip85_entropy"`
			PubHash                                  string `json:"public_key_sha256"`
		}
	}
	require.NoError(t, json.Unmarshal(data, &vectors))
	require.Len(t, vectors.Cases, 9)
	for _, tc := range vectors.Cases {
		t.Run(tc.Path+"/"+tc.Passphrase, func(t *testing.T) {
			path, err := hd.CreateMlDsa65BIP85Path(tc.CoinType, tc.Account, tc.Index)
			require.NoError(t, err)
			require.Equal(t, tc.Path, path)
			bip39Seed, err := hex.DecodeString(tc.Seed)
			require.NoError(t, err)
			master, chain := hd.ComputeMastersFromSeed(bip39Seed)
			entropy, err := hd.DeriveBIP85Entropy(master, chain, path, 64)
			require.NoError(t, err)
			require.Equal(t, tc.Entropy, hex.EncodeToString(entropy))
			seed, err := hd.MlDsa65BIP85.Derive()(tc.Mnemonic, tc.Passphrase, path)
			require.NoError(t, err)
			require.Equal(t, tc.Output, hex.EncodeToString(seed))
			hash := sha256.Sum256(hd.MlDsa65BIP85.Generate()(seed).PubKey().Bytes())
			require.Equal(t, tc.PubHash, hex.EncodeToString(hash[:]))
		})
	}
}

func TestMlDsa65BIP85Vectors(t *testing.T) {
	// Additional Cosmos and boundary vectors generated independently with
	// @scure/bip39 2.0.1, @scure/bip32 2.0.1, Node HMAC-SHA512/SHAKE256,
	// and @noble/post-quantum 0.6.1. SHAKE256 uses the full 64-byte HMAC.
	mnemonic := strings.Repeat("abandon ", 23) + "art"
	for _, tc := range []struct {
		passphrase, path, seed, pubHash string
	}{
		{"", "", "8919898e02b4fe61751c181c910b82be9a2b9f87acc63a73dc9d03ce4930af75", "35d464b10e832cce796da2b2759476363e502576c992a44f621f6b8a367b856a"},
		{"TREZOR", "m/83696968'/505153'/118'/201'/7'/9'", "1d352f561a297f3f6e74e25c5424e5eaf9f946db87a8f7aa8be192a9211f5780", "f2419a2b2a1f6cd4a35117e40a039ad549ef7f6086d1a1651c82ba6d547e0730"},
		{"café", "", "5fb50c6b2e793028402070ba7989fdae8d5943a81ac6610907987d2577993f18", "e492a954966a6941acf0361fd1e338f1120fda7dc0db4a0583e5fe50d4d8c410"},
		{"cafe\u0301", "", "5fb50c6b2e793028402070ba7989fdae8d5943a81ac6610907987d2577993f18", "e492a954966a6941acf0361fd1e338f1120fda7dc0db4a0583e5fe50d4d8c410"},
		{" \tcafé\u00a0 ", "", "beca87407dc5d555b6d33da1d6cb2e05efb2f4b5d3c92c08ac0c3871376035c2", "68f42cfb1c24865f52be3b3a49b964cd850c702b5a6d7ee0924363e55e55c3ac"},
		{" ", "", "2feb9c8d065391d0d9c2f4b4f52cc7cc5f8249e3ca447099a391d058220da265", "1c96be005bca83dc53d635f1d5790b38d4afe5a77b866f9fc29f5de963db92f4"},
		{"", "m/83696968'/505153'/2147483647'/201'/2147483647'/2147483647'", "aaffc11c254080e99800d97d658f406a7a641fa9ca7921d3e574f77a64cd588f", "c3e82f1262af99b8ba4999b3189849b6d1d07bd222326cce679352a62c03547c"},
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
	require.Equal(t, "8919898e02b4fe61751c181c910b82be9a2b9f87acc63a73dc9d03ce4930af75", hex.EncodeToString(seed))
	spaced := "\n\u2003" + strings.ReplaceAll(strings.ReplaceAll(mnemonic, "a", "\uff41"), " ", "\t\u00a0 ") + "\r\n"
	normalized, err := hd.MlDsa65BIP85.Derive()(spaced, "", path)
	require.NoError(t, err)
	require.Equal(t, seed, normalized)
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
		"m/83696968'/505153'/118'/3'/0'/0'",                                          // obsolete pre-family-block draft
		"m/83696968'/505153'/118'/200'/0'/0'", "m/83696968'/505153'/118'/202'/0'/0'", // other ML-DSA variants
		"m/83696968'/505153'/118'/400'/0'/0'", "m/83696968'/505153'/118'/401'/0'/0'", // reserved FN-DSA
		"m/83696968'/505153'/118/201'/0'/0'", "m/83696968'/505153'/118'/201'/0'/0",
		"m/83696968'/505153'/118'/201'/0'/0'/1'", "m/83696968'/505153'/118'/201'/0'",
		"m/83696968'/505153'/2147483648'/201'/0'/0'",
		"m/83696968'/505153'/118'/201'/2147483648'/0'",
		"m/83696968'/505153'/118'/201'/0'/2147483648'",
	} {
		_, err := hd.MlDsa65BIP85.Derive()(mnemonic, "", path)
		require.Error(t, err)
	}
	for _, scope := range [][3]uint32{{1 << 31, 0, 0}, {118, 1 << 31, 0}, {118, 0, 1 << 31}} {
		_, err := hd.CreateMlDsa65BIP85Path(scope[0], scope[1], scope[2])
		require.Error(t, err)
	}
}
