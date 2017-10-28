package cryptostore_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmn "github.com/tendermint/tmlibs/common"

	"github.com/tendermint/go-crypto/keys/cryptostore"
)

func TestNoopEncoder(t *testing.T) {
	assert, require := assert.New(t), require.New(t)
	noop := cryptostore.Noop

	key, err := cryptostore.GenEd25519.Generate(cmn.RandBytes(16))
	require.NoError(err)
	key2, err := cryptostore.GenSecp256k1.Generate(cmn.RandBytes(16))
	require.NoError(err)

	b, err := noop.Encrypt(key, "encode")
	require.Nil(err)
	assert.NotEmpty(b)

	b2, err := noop.Encrypt(key2, "encode")
	require.Nil(err)
	assert.NotEmpty(b2)
	assert.NotEqual(b, b2)

	// note the decode with a different password works - not secure!
	pk, err := noop.Decrypt(b, "decode")
	require.Nil(err)
	require.NotNil(pk)
	assert.Equal(key, pk)

	pk2, err := noop.Decrypt(b2, "kggugougp")
	require.Nil(err)
	require.NotNil(pk2)
	assert.Equal(key2, pk2)
}

func TestSecretBox(t *testing.T) {
	assert, require := assert.New(t), require.New(t)
	enc := cryptostore.SecretBox

	key, err := cryptostore.GenEd25519.Generate(cmn.RandBytes(16))
	require.NoError(err)
	pass := "some-special-secret"

	b, err := enc.Encrypt(key, pass)
	require.Nil(err)
	assert.NotEmpty(b)

	// decoding with a different pass is an error
	pk, err := enc.Decrypt(b, "decode")
	require.NotNil(err)
	require.True(pk.Empty())

	// but decoding with the same passphrase gets us our key
	pk, err = enc.Decrypt(b, pass)
	require.Nil(err)
	assert.Equal(key, pk)
}

func TestSecretBoxNoPass(t *testing.T) {
	assert, require := assert.New(t), require.New(t)
	enc := cryptostore.SecretBox

	key, rerr := cryptostore.GenEd25519.Generate(cmn.RandBytes(16))
	require.NoError(rerr)

	cases := []struct {
		encode string
		decode string
		valid  bool
	}{
		{"foo", "foo", true},
		{"foo", "food", false},
		{"", "", true},
		{"", "a", false},
		{"a", "", false},
	}

	for i, tc := range cases {
		b, err := enc.Encrypt(key, tc.encode)
		require.Nil(err, "%d: %+v", i, err)
		assert.NotEmpty(b, "%d", i)

		pk, err := enc.Decrypt(b, tc.decode)
		if tc.valid {
			require.Nil(err, "%d: %+v", i, err)
			assert.Equal(key, pk, "%d", i)
		} else {
			require.NotNil(err, "%d", i)
		}
	}

	// now let's make sure raw bytes also work...
	b := key.Bytes()
	pk, err := enc.Decrypt(b, "")
	require.Nil(err, "%+v", err)
	assert.Equal(key, pk)
}
