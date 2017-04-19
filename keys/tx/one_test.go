package tx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	crypto "github.com/tendermint/go-crypto"
	keys "github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/go-crypto/keys/cryptostore"
	"github.com/tendermint/go-crypto/keys/storage/memstorage"
)

func TestOneSig(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	algo := crypto.NameEd25519
	cstore := cryptostore.New(
		cryptostore.SecretBox,
		memstorage.New(),
	)
	n, p := "foo", "bar"
	n2, p2 := "other", "thing"

	acct, err := cstore.Create(n, p, algo)
	require.Nil(err, "%+v", err)
	acct2, err := cstore.Create(n2, p2, algo)
	require.Nil(err, "%+v", err)

	cases := []struct {
		data       string
		key        keys.Info
		name, pass string
	}{
		{"first", acct, n, p},
		{"kehfkhefy8y", acct, n, p},
		{"second", acct2, n2, p2},
	}

	for _, tc := range cases {
		tx := New([]byte(tc.data))
		// unsigned version
		_, err = tx.Signers()
		assert.NotNil(err)
		orig, err := tx.TxBytes()
		require.Nil(err, "%+v", err)
		data := tx.SignBytes()
		assert.Equal(tc.data, string(data))

		// sign it
		err = cstore.Sign(tc.name, tc.pass, tx)
		require.Nil(err, "%+v", err)
		// but not twice
		err = cstore.Sign(tc.name, tc.pass, tx)
		require.NotNil(err)

		// make sure it is proper now
		sigs, err := tx.Signers()
		require.Nil(err, "%+v", err)
		if assert.Equal(1, len(sigs)) {
			// This must be refactored...
			assert.Equal(tc.key.PubKey, sigs[0])
		}
		// the tx bytes should change after this
		after, err := tx.TxBytes()
		require.Nil(err, "%+v", err)
		assert.NotEqual(orig, after, "%X != %X", orig, after)

		// sign bytes are the same
		data = tx.SignBytes()
		assert.Equal(tc.data, string(data))
	}
}
