package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkkv "github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding connection type.
func NewDecodeStore(cdc codec.BinaryMarshaler, kvA, kvB sdkkv.Pair) (string, bool) {
	switch {
	case bytes.HasPrefix(kvA.Key, host.KeyClientStorePrefix) && bytes.HasSuffix(kvA.Key, host.KeyConnectionPrefix):
		var clientConnectionsA, clientConnectionsB types.ClientPaths
		cdc.MustUnmarshalBinaryBare(kvA.Value, &clientConnectionsA)
		cdc.MustUnmarshalBinaryBare(kvB.Value, &clientConnectionsB)
		return fmt.Sprintf("ClientPaths A: %v\nClientPaths B: %v", clientConnectionsA, clientConnectionsB), true

	case bytes.HasPrefix(kvA.Key, host.KeyConnectionPrefix):
		var connectionA, connectionB types.ConnectionEnd
		cdc.MustUnmarshalBinaryBare(kvA.Value, &connectionA)
		cdc.MustUnmarshalBinaryBare(kvB.Value, &connectionB)
		return fmt.Sprintf("ConnectionEnd A: %v\nConnectionEnd B: %v", connectionA, connectionB), true

	default:
		return "", false
	}
}
