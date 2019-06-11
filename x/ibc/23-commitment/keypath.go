package commitment

import (
	"github.com/tendermint/tendermint/crypto/merkle"
)

// Hard coded for now
func SDKPrefix() merkle.KeyPath {
	return new(merkle.KeyPath).
		AppendKey([]byte("ibc"), merkle.KeyEncodingHex).
		AppendKey([]byte{0x00}, merkle.KeyEncodingHex)
}

func PrefixKeyPath(prefix string, key []byte) (res merkle.KeyPath, err error) {
	keys, err := merkle.KeyPathToKeys(prefix)
	if err != nil {
		return
	}

	keys[len(keys)-1] = append(keys[len(keys)], key...)

	for _, key := range keys {
		res = res.AppendKey(key, merkle.KeyEncodingHex)
	}

	return
}
