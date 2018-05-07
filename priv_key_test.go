package crypto_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	crypto "github.com/tendermint/go-crypto"
)

func TestGeneratePrivKey(t *testing.T) {
	testPriv := crypto.GenPrivKeyEd25519()
	testGenerate := testPriv.Generate(1)
	signBytes := []byte("something to sign")
	assert.True(t, testGenerate.PubKey().VerifyBytes(signBytes, testGenerate.Sign(signBytes)))
}

/*

type BadKey struct {
	PrivKeyEd25519
}

func TestReadPrivKey(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	// garbage in, garbage out
	garbage := []byte("hjgewugfbiewgofwgewr")
	XXX This test wants to register BadKey globally to go-crypto,
	but we don't want to support that.
	_, err := PrivKeyFromBytes(garbage)
	require.Error(err)

	edKey := GenPrivKeyEd25519()
	badKey := BadKey{edKey}

	cases := []struct {
		key   PrivKey
		valid bool
	}{
		{edKey, true},
		{badKey, false},
	}

	for i, tc := range cases {
		data := tc.key.Bytes()
		fmt.Println(">>>", data)
		key, err := PrivKeyFromBytes(data)
		fmt.Printf("!!! %#v\n", key, err)
		if tc.valid {
			assert.NoError(err, "%d", i)
			assert.Equal(tc.key, key, "%d", i)
		} else {
			assert.Error(err, "%d: %#v", i, key)
		}
	}
}
*/
