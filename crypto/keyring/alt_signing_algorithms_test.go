package keyring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAltSigningAlgoList_Contains(t *testing.T) {
	list := AltSigningAlgoList{
		AltSecp256k1,
	}

	notSupportedAlgo := AltSigningAlgo{
		Name:      "anotherAlgo",
		DeriveKey: nil,
	}
	assert.True(t, list.Contains(AltSecp256k1))
	assert.False(t, list.Contains(notSupportedAlgo))
}
