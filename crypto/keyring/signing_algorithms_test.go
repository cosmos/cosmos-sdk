package keyring

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/algo"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestDefaults(t *testing.T) {
	require.Equal(t, algo.PubKeyType("multi"), algo.MultiAlgo)
	require.Equal(t, algo.PubKeyType("secp256k1"), algo.Secp256k1)
	require.Equal(t, algo.PubKeyType("ed25519"), algo.Ed25519)
	require.Equal(t, algo.PubKeyType("sr25519"), algo.Sr25519)
}

func TestNewSigningAlgoByString(t *testing.T) {
	tests := []struct {
		name         string
		algoStr      string
		isSupported  bool
		expectedAlgo SignatureAlgo
		expectedErr  error
	}{
		{
			"supported algorithm",
			"secp256k1",
			true,
			algo.AltSecp256k1,
			nil,
		},
		{
			"not supported",
			"notsupportedalgo",
			false,
			nil,
			fmt.Errorf("provided algorithm `notsupportedalgo` is not supported"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			algorithm, err := NewSigningAlgoFromString(tt.algoStr)
			if tt.isSupported {
				require.Equal(t, algo.AltSecp256k1, algorithm)
			} else {
				require.EqualError(t, err, tt.expectedErr.Error())
			}
		})
	}
}

func TestAltSigningAlgoList_Contains(t *testing.T) {
	list := SigningAlgoList{
		algo.AltSecp256k1,
	}

	assert.True(t, list.Contains(algo.AltSecp256k1))
	assert.False(t, list.Contains(notSupportedAlgo{}))
}

type notSupportedAlgo struct {
}

func (n notSupportedAlgo) Name() algo.PubKeyType {
	return "notSupported"
}

func (n notSupportedAlgo) DeriveKey() algo.DeriveKeyFn {
	return algo.Secp256k1DeriveKey
}

func (n notSupportedAlgo) PrivKeyGen() algo.PrivKeyGenFn {
	return algo.Secp256k1PrivKeyGen
}
