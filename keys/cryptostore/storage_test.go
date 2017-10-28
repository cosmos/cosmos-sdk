package cryptostore

import (
	"testing"

	"github.com/stretchr/testify/assert"

	crypto "github.com/tendermint/go-crypto"
	cmn "github.com/tendermint/tmlibs/common"

	keys "github.com/tendermint/go-crypto/keys"
)

func TestSortKeys(t *testing.T) {
	assert := assert.New(t)

	gen := func() crypto.PrivKey {
		key, _ := GenEd25519.Generate(cmn.RandBytes(16))
		return key
	}
	assert.NotEqual(gen(), gen())

	// alphabetical order is n3, n1, n2
	n1, n2, n3 := "john", "mike", "alice"
	infos := keys.Infos{
		info(n1, gen()),
		info(n2, gen()),
		info(n3, gen()),
	}

	// make sure they are initialized unsorted
	assert.Equal(n1, infos[0].Name)
	assert.Equal(n2, infos[1].Name)
	assert.Equal(n3, infos[2].Name)

	// now they are sorted
	infos.Sort()
	assert.Equal(n3, infos[0].Name)
	assert.Equal(n1, infos[1].Name)
	assert.Equal(n2, infos[2].Name)

	// make sure info put some real data there...
	assert.NotEmpty(infos[0].PubKey)
	assert.NotEmpty(infos[0].PubKey.Address())
	assert.NotEmpty(infos[1].PubKey)
	assert.NotEmpty(infos[1].PubKey.Address())
	assert.NotEqual(infos[0].PubKey, infos[1].PubKey)
}
