package keyring

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/privkey"
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
			privkey.Secp256k1,
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
				require.Equal(t, privkey.Secp256k1, algorithm)
			} else {
				require.EqualError(t, err, tt.expectedErr.Error())
			}
		})
	}
}

func TestAltSigningAlgoList_Contains(t *testing.T) {
	list := SigningAlgoList{
		privkey.Secp256k1,
	}

	assert.True(t, list.Contains(privkey.Secp256k1))
	assert.False(t, list.Contains(notSupportedAlgo{}))
}

type notSupportedAlgo struct {
}

func (n notSupportedAlgo) Name() privkey.PubKeyType {
	return "notSupported"
}

func (n notSupportedAlgo) Derive() privkey.DeriveKeyFn {
	return privkey.Secp256k1.Derive()
}

func (n notSupportedAlgo) Generate() privkey.GenerateFn {
	return privkey.Secp256k1.Generate()
}
