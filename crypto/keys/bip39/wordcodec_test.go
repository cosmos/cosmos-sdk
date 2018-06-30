package bip39

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWordCodec_NewMnemonic(t *testing.T) {
	_, err := NewMnemonic(FundRaiser)
	require.NoError(t, err, "unexpected error generating fundraiser mnemonic")

	_, err = NewMnemonic(FreshKey)
	require.NoError(t, err, "unexpected error generating new 24-word mnemonic")
}
