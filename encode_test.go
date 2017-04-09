package crypto

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	data "github.com/tendermint/go-data"
	wire "github.com/tendermint/go-wire"
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

// make sure go-wire json can still figure this out...
func checkWireJSON(t *testing.T, in interface{}, reader interface{}, typ byte) {
	// test to and from binary
	var err error
	js := wire.JSONBytes(in)
	btyp := fmt.Sprintf("[%d,", typ)
	assert.True(t, strings.HasPrefix(string(js), btyp), string(js))

	wire.ReadJSON(reader, js, &err)
	require.Nil(t, err, "%+v", err)
}

func TestKeyEncodings(t *testing.T) {
	cases := []struct {
		privKey PrivKey
		keyType byte
		keyName string
	}{
		{
			privKey: GenPrivKeyEd25519().Wrap(),
			keyType: TypeEd25519,
			keyName: NameEd25519,
		},
		{
			privKey: GenPrivKeySecp256k1().Wrap(),
			keyType: TypeSecp256k1,
			keyName: NameSecp256k1,
		},
	}

	for _, tc := range cases {
		// check (de/en)codings of private key
		var priv2, priv3, priv4 PrivKey
		checkWire(t, tc.privKey, &priv2, tc.keyType)
		assert.EqualValues(t, tc.privKey, priv2)
		checkJSON(t, tc.privKey, &priv3, tc.keyName)
		assert.EqualValues(t, tc.privKey, priv3)
		checkWireJSON(t, tc.privKey, &priv4, tc.keyType)
		assert.EqualValues(t, tc.privKey, priv4)

		// check (de/en)codings of public key
		pubKey := tc.privKey.PubKey()
		var pub2, pub3, pub4 PubKey
		checkWire(t, pubKey, &pub2, tc.keyType)
		assert.EqualValues(t, pubKey, pub2)
		checkJSON(t, pubKey, &pub3, tc.keyName)
		assert.EqualValues(t, pubKey, pub3)
		checkWireJSON(t, pubKey, &pub4, tc.keyType)
		assert.EqualValues(t, pubKey, pub4)
	}
}

func toFromJSON(t *testing.T, in interface{}, recvr interface{}) {
	js, err := data.ToJSON(in)
	require.Nil(t, err, "%+v", err)
	err = data.FromJSON(js, recvr)
	require.Nil(t, err, "%+v", err)
}

func TestNilEncodings(t *testing.T) {
	// make sure sigs are okay with nil
	var a, b Signature
	toFromJSON(t, a, &b)
	assert.EqualValues(t, a, b)

	// make sure sigs are okay with nil
	var c, d PubKey
	toFromJSON(t, c, &d)
	assert.EqualValues(t, c, d)

	// make sure sigs are okay with nil
	var e, f PrivKey
	toFromJSON(t, e, &f)
	assert.EqualValues(t, e, f)

}
