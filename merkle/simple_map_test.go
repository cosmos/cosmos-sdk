package merkle

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type strHasher string

func (str strHasher) Hash() []byte {
	return SimpleHashFromBytes([]byte(str))
}

func TestSimpleMap(t *testing.T) {
	{
		db := NewSimpleMap()
		db.Set("key1", strHasher("value1"))
		assert.Equal(t, "acdb4f121bc6f25041eb263ab463f1cd79236a32", fmt.Sprintf("%x", db.Hash()), "Hash didn't match")
	}
	{
		db := NewSimpleMap()
		db.Set("key1", strHasher("value2"))
		assert.Equal(t, "b8cbf5adee8c524e14f531da9b49adbbbd66fffa", fmt.Sprintf("%x", db.Hash()), "Hash didn't match")
	}
	{
		db := NewSimpleMap()
		db.Set("key1", strHasher("value1"))
		db.Set("key2", strHasher("value2"))
		assert.Equal(t, "1708aabc85bbe00242d3db8c299516aa54e48c38", fmt.Sprintf("%x", db.Hash()), "Hash didn't match")
	}
	{
		db := NewSimpleMap()
		db.Set("key2", strHasher("value2")) // NOTE: out of order
		db.Set("key1", strHasher("value1"))
		assert.Equal(t, "1708aabc85bbe00242d3db8c299516aa54e48c38", fmt.Sprintf("%x", db.Hash()), "Hash didn't match")
	}
	{
		db := NewSimpleMap()
		db.Set("key1", strHasher("value1"))
		db.Set("key2", strHasher("value2"))
		db.Set("key3", strHasher("value3"))
		assert.Equal(t, "e728afe72ce351eed6aca65c5f78da19b9a6e214", fmt.Sprintf("%x", db.Hash()), "Hash didn't match")
	}
	{
		db := NewSimpleMap()
		db.Set("key2", strHasher("value2")) // NOTE: out of order
		db.Set("key1", strHasher("value1"))
		db.Set("key3", strHasher("value3"))
		assert.Equal(t, "e728afe72ce351eed6aca65c5f78da19b9a6e214", fmt.Sprintf("%x", db.Hash()), "Hash didn't match")
	}
}
