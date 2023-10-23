package types

import (
	"io"

	"github.com/bcicen/jstream"
)

const ChainIDFieldName = "chain-id"

// ParseChainIDFromGenesis parses the chain-id from the genesis file using constant memory.
func ParseChainIDFromGenesis(reader io.Reader) (string, error) {
	decoder := jstream.NewDecoder(reader, 1).EmitKV()
	for mv := range decoder.Stream() {
		kv := mv.Value.(jstream.KV)
		if kv.Key == ChainIDFieldName {
			if chain_id, ok := kv.Value.(string); ok {
				return chain_id, nil
			}
			break
		}
	}
	return "", nil
}
