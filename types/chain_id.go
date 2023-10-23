package types

import (
	"errors"
	"io"

	"github.com/bcicen/jstream"
)

const ChainIDFieldName = "chain_id"

// ParseChainIDFromGenesis parses the chain-id from the genesis file using constant memory.
func ParseChainIDFromGenesis(reader io.Reader) (string, error) {
	decoder := jstream.NewDecoder(reader, 1).EmitKV()
	var chain_id string
	err := decoder.Decode(func(mv *jstream.MetaValue) bool {
		if kv, ok := mv.Value.(jstream.KV); ok {
			if kv.Key == ChainIDFieldName {
				chain_id, _ = kv.Value.(string)
				return false
			}
		}
		return true
	})
	if len(chain_id) > 0 {
		return chain_id, nil
	}
	if err == nil {
		return "", errors.New("chain-id not found in genesis file")
	}
	return "", err
}
