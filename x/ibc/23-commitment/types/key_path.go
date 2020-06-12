package types

import (
	fmt "fmt"
	"net/url"
)

// AppendKey appends a new key to a KeyPath
func (pth KeyPath) AppendKey(key []byte, enc KeyEncoding) KeyPath {
	pth.Keys = append(pth.Keys, &Key{name: key, enc: enc})
	return pth
}

// String implements the fmt.Stringer interface
func (pth *KeyPath) String() string {
	res := ""
	for _, key := range pth.Keys {
		switch key.enc {
		case URL:
			res += "/" + url.PathEscape(string(key.name))
		case HEX:
			res += "/x:" + fmt.Sprintf("%X", key.name)
		default:
			panic("unexpected key encoding type")
		}
	}
	return res
}

// GetKey returns the bytes representation of key at given index
// Passing in a positive index return the key at index in forward order
// from the highest key to the lowest key
// Passing in a negative index will return the key at index in reverse order
// from the lowest key to the highest key. This is the order for proof verification,
// since we prove lowest key first before moving to key of higher subtrees
func (pth *KeyPath) GetKey(i int) []byte {
	total := len(pth.Keys)
	index := (total + i) % total
	return pth.Keys[index].name
}
