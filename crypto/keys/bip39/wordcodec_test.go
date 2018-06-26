package bip39

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWordCodec_NewMnemonic(t *testing.T) {
	_, err := NewMnemonic(FundRaiser)
	assert.NoError(t, err, "unexpected error generating fundraiser mnemonic")

	_, err = NewMnemonic(FreshKey)
	assert.NoError(t, err, "unexpected error generating new 24-word mnemonic")
}
