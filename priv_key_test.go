package crypto

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wire "github.com/tendermint/go-wire"
)

type BadKey struct {
	PrivKeyEd25519
}

// Wrap fulfils interface for PrivKey struct
func (pk BadKey) Wrap() PrivKey {
	return PrivKey{pk}
}

func (pk BadKey) Bytes() []byte {
	return wire.BinaryBytes(pk.Wrap())
}

func (pk BadKey) ValidateKey() error {
	return fmt.Errorf("fuggly key")
}

func init() {
	PrivKeyMapper.
		RegisterImplementation(BadKey{}, "bad", 0x66)
}

func TestReadPrivKey(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	// garbage in, garbage out
	garbage := []byte("hjgewugfbiewgofwgewr")
	_, err := PrivKeyFromBytes(garbage)
	require.Error(err)

	edKey := GenPrivKeyEd25519()
	badKey := BadKey{edKey}

	cases := []struct {
		key   PrivKey
		valid bool
	}{
		{edKey.Wrap(), true},
		{badKey.Wrap(), false},
	}

	for i, tc := range cases {
		data := tc.key.Bytes()
		key, err := PrivKeyFromBytes(data)
		if tc.valid {
			assert.NoError(err, "%d", i)
			assert.Equal(tc.key, key, "%d", i)
		} else {
			assert.Error(err, "%d: %#v", i, key)
		}
	}

}
