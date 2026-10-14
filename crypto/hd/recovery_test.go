package hd_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/types"
)

func TestAccountKeyDerivations(t *testing.T) {
	derivations, err := hd.AccountKeyDerivations(118, 7, 9)
	require.NoError(t, err)
	require.Len(t, derivations, 3)

	require.Equal(t, "pqhd-201", derivations[0].Scheme)
	require.Equal(t, hd.MlDsa65Type, derivations[0].PubKeyType)
	require.Equal(t, "m/83696968'/505153'/118'/201'/7'/9'", derivations[0].Path)
	require.Equal(t, "bip44", derivations[1].Scheme)
	require.Equal(t, hd.Secp256k1Type, derivations[1].PubKeyType)
	require.Equal(t, "m/44'/118'/7'/0/9", derivations[1].Path)
	require.Equal(t, "bip44", derivations[2].Scheme)
	require.Equal(t, hd.MlDsa65Type, derivations[2].PubKeyType)
	require.Equal(t, "m/44'/118'/7'/0/9", derivations[2].Path)

	mnemonic := strings.Repeat("abandon ", 23) + "art"
	seen := map[string]bool{}
	for _, d := range derivations {
		bz, err := d.Algo.Derive()(mnemonic, "", d.Path)
		require.NoError(t, err)
		pub := d.Algo.Generate()(bz).PubKey()
		require.Equal(t, string(d.PubKeyType), pub.Type())
		require.False(t, seen[string(pub.Bytes())], "%s %s", d.Scheme, d.Path)
		seen[string(pub.Bytes())] = true
	}

	bz, err := hd.MlDsa65BIP85.Derive()(mnemonic, "", derivations[0].Path)
	require.NoError(t, err)
	expected := hd.MlDsa65BIP85.Generate()(bz).PubKey()
	bz, err = derivations[0].Algo.Derive()(mnemonic, "", derivations[0].Path)
	require.NoError(t, err)
	require.True(t, expected.Equals(derivations[0].Algo.Generate()(bz).PubKey()))

	_, err = hd.AccountKeyDerivations(118, 1<<31, 0)
	require.ErrorContains(t, err, "must be less than")
}

// restoreCandidates is the restore loop documented in the keyring guide.
func restoreCandidates(mnemonic string) ([]types.PubKey, error) {
	derivations, err := hd.AccountKeyDerivations(118, 0, 0)
	if err != nil {
		return nil, err
	}
	var pubKeys []types.PubKey
	for _, d := range derivations {
		bz, err := d.Algo.Derive()(mnemonic, "", d.Path)
		if errors.Is(err, hd.ErrMlDsa65BIP85Mnemonic) {
			continue // PQHD needs 24 words; shorter backups only have BIP44 keys.
		}
		if err != nil {
			return nil, err
		}
		pubKeys = append(pubKeys, d.Algo.Generate()(bz).PubKey())
	}
	return pubKeys, nil
}

func TestAccountKeyDerivationsRestoreLoop(t *testing.T) {
	short := strings.Repeat("abandon ", 11) + "about"
	pubKeys, err := restoreCandidates(short)
	require.NoError(t, err)
	require.Len(t, pubKeys, 2)
	for i, algo := range []hd.AccountKeyAlgo{hd.Secp256k1, hd.MlDsa65} {
		bz, err := algo.Derive()(short, "", "m/44'/118'/0'/0/0")
		require.NoError(t, err)
		require.True(t, algo.Generate()(bz).PubKey().Equals(pubKeys[i]))
	}

	pubKeys, err = restoreCandidates(strings.Repeat("abandon ", 23) + "art")
	require.NoError(t, err)
	require.Len(t, pubKeys, 3)

	_, err = restoreCandidates(strings.Repeat("abandon ", 12))
	require.Error(t, err)
	require.NotErrorIs(t, err, hd.ErrMlDsa65BIP85Mnemonic)
	_, err = restoreCandidates(strings.Repeat("abandon ", 24))
	require.Error(t, err)
	require.NotErrorIs(t, err, hd.ErrMlDsa65BIP85Mnemonic)
}
