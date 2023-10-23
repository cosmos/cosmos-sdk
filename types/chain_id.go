package types

import (
	"errors"
	fmt "fmt"
	"io"

	"github.com/bcicen/jstream"
)

const ChainIDFieldName = "chain-id"

// ParseChainIDFromGenesis parses the chain-id from the genesis file using constant memory.
func ParseChainIDFromGenesis(reader io.Reader) (string, error) {
	decoder := jstream.NewDecoder(reader, 1).EmitKV()
	for mv := range decoder.Stream() {
		if kv, ok := mv.Value.(jstream.KV); ok {
			if kv.Key == ChainIDFieldName {
				chain_id, ok := kv.Value.(string)
				if !ok {
					return "", fmt.Errorf("chain-id field is not string")
				}
				return chain_id, nil
			}
		}
	}
	return "", errors.New("chain-id field not found")
}
