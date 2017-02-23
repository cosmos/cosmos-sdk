package crypto

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	data "github.com/tendermint/go-data"
)

type byter interface {
	Bytes() []byte
}

// go to wire encoding and back
func checkWire(t *testing.T, in byter, reader interface{}, typ byte) {
	// test to and from binary
	bin, err := data.ToWire(in)
	require.Nil(t, err, "%+v", err)
	assert.Equal(t, typ, bin[0])
	// make sure this is compatible with current (Bytes()) encoding
	assert.Equal(t, in.Bytes(), bin)

	err = data.FromWire(bin, reader)
	require.Nil(t, err, "%+v", err)
}

// go to json encoding and back
func checkJSON(t *testing.T, in interface{}, reader interface{}, typ string) {
	// test to and from binary
	js, err := data.ToJSON(in)
	require.Nil(t, err, "%+v", err)
	styp := `"` + typ + `"`
	assert.True(t, strings.Contains(string(js), styp))

	err = data.FromJSON(js, reader)
	require.Nil(t, err, "%+v", err)

	// also check text format
	text, err := data.ToText(in)
	require.Nil(t, err, "%+v", err)
	parts := strings.Split(text, ":")
	require.Equal(t, 2, len(parts))
	// make sure the first part is the typ string
	assert.Equal(t, typ, parts[0])
	// and the data is also present in the json
	assert.True(t, strings.Contains(string(js), parts[1]))
}

func TestKeyEncodings(t *testing.T) {
	cases := []struct {
		privKey PrivKeyS
		keyType byte
		keyName string
	}{
		{
			privKey: PrivKeyS{GenPrivKeyEd25519()},
			keyType: PrivKeyTypeEd25519,
			keyName: PrivKeyNameEd25519,
		},
		{
			privKey: PrivKeyS{GenPrivKeySecp256k1()},
			keyType: PrivKeyTypeSecp256k1,
			keyName: PrivKeyNameSecp256k1,
		},
	}

	for _, tc := range cases {
		// check (de/en)codings of private key
		priv2 := PrivKeyS{}
		checkWire(t, tc.privKey, &priv2, tc.keyType)
		assert.EqualValues(t, tc.privKey, priv2)
		priv3 := PrivKeyS{}
		checkJSON(t, tc.privKey, &priv3, tc.keyName)
		assert.EqualValues(t, tc.privKey, priv3)

		// check (de/en)codings of public key
		pubKey := PubKeyS{tc.privKey.PubKey()}
		pub2 := PubKeyS{}
		checkWire(t, pubKey, &pub2, tc.keyType)
		assert.EqualValues(t, pubKey, pub2)
		pub3 := PubKeyS{}
		checkJSON(t, pubKey, &pub3, tc.keyName)
		assert.EqualValues(t, pubKey, pub3)
	}

}
