package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	crypto "github.com/tendermint/go-crypto"
)

func TestHexData(t *testing.T) {
	assert := assert.New(t)

	orig := HexData("!@##sdfg")
	enc, err := json.Marshal(orig)
	assert.Nil(err)

	var parsed HexData
	err = json.Unmarshal(enc, &parsed)
	assert.Nil(err)

	assert.Equal(orig, parsed)

	bin := HexData{79, 3, 72, 4}
	be, err := json.Marshal(bin)
	assert.Nil(err)
	// make sure proper hex
	assert.Equal(be[0], byte('"'))
	assert.Equal(be[1], byte('4'))
	assert.Equal(be[2], byte('f'))

	err = json.Unmarshal(be, &parsed)
	assert.Nil(err)

	assert.Equal(bin, parsed)
}

func TestPubKey(t *testing.T) {
	assert := assert.New(t)

	key := crypto.GenPrivKeyEd25519().PubKey()
	jkey := JSONPubKey{key}
	enc, err := json.Marshal(jkey)
	assert.Nil(err)
	// see it is nice string (always prefix with byte 1 for ed25519)
	assert.Equal(enc[0], byte('"'))
	assert.Equal(enc[1], byte('0'))
	assert.Equal(enc[2], byte('1'))

	var parsed JSONPubKey
	err = json.Unmarshal(enc, &parsed)
	assert.Nil(err)

	assert.Equal(key, parsed.PubKey)
}
