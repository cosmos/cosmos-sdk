package keyring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAltSigningAlgoList_Contains(t *testing.T) {
	list := AltSigningAlgoList{
		AltSecp256k1,
	}

	assert.True(t, list.Contains(AltSecp256k1))
	assert.False(t, list.Contains(notSupportedAlgo{}))
}

type notSupportedAlgo struct {
}

func (n notSupportedAlgo) String() string {
	return "notSupported"
}

func (n notSupportedAlgo) DeriveKey() AltDeriveKeyFunc {
	return SecpDeriveKey
}

func (n notSupportedAlgo) PrivKeyGen() AltPrivKeyGenFunc {
	return SecpPrivKeyGen
}
