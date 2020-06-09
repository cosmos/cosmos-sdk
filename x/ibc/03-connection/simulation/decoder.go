package simulation

import (
	"bytes"
	"fmt"

	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/03-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding connection type.
func NewDecodeStore(cdc codec.Marshaler) func(kvA, kvB tmkv.Pair) string {
	return func(kvA, kvB tmkv.Pair) string {
		switch {
		case bytes.HasPrefix(kvA.Key, host.KeyClientStorePrefix) && bytes.HasSuffix(kvA.Key, host.KeyConnectionPrefix):
			var clientConnectionsA, clientConnectionsB []string
			cdc.MustUnmarshalBinaryBare(kvA.Value, &clientConnectionsA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &clientConnectionsB)
			return fmt.Sprintf("clients connection IDs A: %v\nclients connection IDs B:%v", clientConnectionsA, clientConnectionsB)

		case bytes.HasPrefix(kvA.Key, host.KeyConnectionPrefix):
			var connectionA, connectionB types.ConnectionEnd
			cdc.MustUnmarshalBinaryBare(kvA.Value, &connectionA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &connectionB)
			return fmt.Sprintf("ConnectionEnd A: %d\nConnectionEnd B: %d\n", seqA, seqB)

		default:
			panic(fmt.Sprintf("invalid %s %s key prefix: %s", host.ModuleName, types.SubModuleName, string(kvA.Key)))
		}
	}
}
