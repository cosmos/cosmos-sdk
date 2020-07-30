package bech32_test

import (
	"bytes"
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types/bech32"
)

func TestEncodeAndDecode(t *testing.T) {
	sum := sha256.Sum256([]byte("hello world\n"))
	ss := "shasum"

	bech, err := bech32.ConvertAndEncode(ss, sum[:])
	require.NoError(t, err)

	hrp, data, err := bech32.DecodeAndConvert(bech)
	require.NoError(t, err)

	require.Equal(t, hrp, ss, "Invalid hrp")
	require.True(t, bytes.Equal(data, sum[:]), "Invalid decode")
}
