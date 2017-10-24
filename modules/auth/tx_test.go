package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	crypto "github.com/tendermint/go-crypto"
	keys "github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/go-crypto/keys/cryptostore"
	"github.com/tendermint/go-crypto/keys/storage/memstorage"
	wire "github.com/tendermint/go-wire"

	"github.com/cosmos/cosmos-sdk/util"
)

func checkSignBytes(t *testing.T, bytes []byte, expected string) {
	// load it back... unwrap the tx
	var raw util.RawTx
	err := wire.ReadBinaryBytes(bytes, &raw)
	require.Nil(t, err)
	// now make sure this is data.Bytes with the info we expect
	assert.Equal(t, expected, string(raw.Data))
}

func TestOneSig(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	algo := crypto.NameEd25519
	cstore := cryptostore.New(
		cryptostore.SecretBox,
		memstorage.New(),
		keys.MustLoadCodec("english"),
	)
	n, p := "foo", "bar"
	n2, p2 := "other", "thing"

	acct, _, err := cstore.Create(n, p, algo)
	require.Nil(err, "%+v", err)
	acct2, _, err := cstore.Create(n2, p2, algo)
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
		tx := newSingle([]byte(tc.data))
		// unsigned version
		_, err = tx.Signers()
		assert.NotNil(err)
		orig, err := tx.TxBytes()
		require.Nil(err, "%+v", err)
		data := tx.SignBytes()
		checkSignBytes(t, data, tc.data)

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
		checkSignBytes(t, data, tc.data)
	}
}

func TestMultiSig(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	algo := crypto.NameEd25519
	cstore := cryptostore.New(
		cryptostore.SecretBox,
		memstorage.New(),
		keys.MustLoadCodec("english"),
	)
	n, p := "foo", "bar"
	n2, p2 := "other", "thing"

	acct, _, err := cstore.Create(n, p, algo)
	require.Nil(err, "%+v", err)
	acct2, _, err := cstore.Create(n2, p2, algo)
	require.Nil(err, "%+v", err)

	type signer struct {
		key        keys.Info
		name, pass string
	}
	cases := []struct {
		data    string
		signers []signer
	}{
		{"one", []signer{{acct, n, p}}},
		{"two", []signer{{acct2, n2, p2}}},
		{"both", []signer{{acct, n, p}, {acct2, n2, p2}}},
	}

	for _, tc := range cases {
		tx := newMulti([]byte(tc.data))
		// unsigned version
		_, err = tx.Signers()
		assert.NotNil(err)
		orig, err := tx.TxBytes()
		require.Nil(err, "%+v", err)
		data := tx.SignBytes()
		checkSignBytes(t, data, tc.data)

		// sign it
		for _, s := range tc.signers {
			err = cstore.Sign(s.name, s.pass, tx)
			require.Nil(err, "%+v", err)
		}

		// make sure it is proper now
		sigs, err := tx.Signers()
		require.Nil(err, "%+v", err)
		if assert.Equal(len(tc.signers), len(sigs)) {
			for i := range sigs {
				// This must be refactored...
				assert.Equal(tc.signers[i].key.PubKey, sigs[i])
			}
		}
		// the tx bytes should change after this
		after, err := tx.TxBytes()
		require.Nil(err, "%+v", err)
		assert.NotEqual(orig, after, "%X != %X", orig, after)

		// sign bytes are the same
		data = tx.SignBytes()
		checkSignBytes(t, data, tc.data)
	}
}
