package merkle

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/go-crypto/tmhash"
)

type strHasher string

func (str strHasher) Hash() []byte {
	return tmhash.Sum([]byte(str))
}

func TestSimpleMap(t *testing.T) {
	{
		db := newSimpleMap()
		db.Set("key1", strHasher("value1"))
		assert.Equal(t, "fa9bc106ffd932d919bee935ceb6cf2b3dd72d8f", fmt.Sprintf("%x", db.Hash()), "Hash didn't match")
	}
	{
		db := newSimpleMap()
		db.Set("key1", strHasher("value2"))
		assert.Equal(t, "e00e7dcfe54e9fafef5111e813a587f01ba9c3e8", fmt.Sprintf("%x", db.Hash()), "Hash didn't match")
	}
	{
		db := newSimpleMap()
		db.Set("key1", strHasher("value1"))
		db.Set("key2", strHasher("value2"))
		assert.Equal(t, "eff12d1c703a1022ab509287c0f196130123d786", fmt.Sprintf("%x", db.Hash()), "Hash didn't match")
	}
	{
		db := newSimpleMap()
		db.Set("key2", strHasher("value2")) // NOTE: out of order
		db.Set("key1", strHasher("value1"))
		assert.Equal(t, "eff12d1c703a1022ab509287c0f196130123d786", fmt.Sprintf("%x", db.Hash()), "Hash didn't match")
	}
	{
		db := newSimpleMap()
		db.Set("key1", strHasher("value1"))
		db.Set("key2", strHasher("value2"))
		db.Set("key3", strHasher("value3"))
		assert.Equal(t, "b2c62a277c08dbd2ad73ca53cd1d6bfdf5830d26", fmt.Sprintf("%x", db.Hash()), "Hash didn't match")
	}
	{
		db := newSimpleMap()
		db.Set("key2", strHasher("value2")) // NOTE: out of order
		db.Set("key1", strHasher("value1"))
		db.Set("key3", strHasher("value3"))
		assert.Equal(t, "b2c62a277c08dbd2ad73ca53cd1d6bfdf5830d26", fmt.Sprintf("%x", db.Hash()), "Hash didn't match")
	}
}
