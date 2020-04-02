package keyring

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

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
			AltSecp256k1,
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
			algo, err := NewSigningAlgoFromString(tt.algoStr)
			if tt.isSupported {
				require.Equal(t, AltSecp256k1, algo)
			} else {
				require.EqualError(t, err, tt.expectedErr.Error())
			}
		})
	}
}

func TestAltSigningAlgoList_Contains(t *testing.T) {
	list := SigningAlgoList{
		AltSecp256k1,
	}

	assert.True(t, list.Contains(AltSecp256k1))
	assert.False(t, list.Contains(notSupportedAlgo{}))
}

type notSupportedAlgo struct {
}

func (n notSupportedAlgo) Name() pubKeyType {
	return "notSupported"
}

func (n notSupportedAlgo) DeriveKey() DeriveKeyFn {
	return Secp256k1DeriveKey
}

func (n notSupportedAlgo) PrivKeyGen() PrivKeyGenFn {
	return Secp256k1PrivKeyGen
}
