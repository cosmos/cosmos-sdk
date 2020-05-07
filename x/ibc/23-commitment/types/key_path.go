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
