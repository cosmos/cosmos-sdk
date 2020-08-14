package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/keeper"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var _ ClientUnmarshaler = (*keeper.Keeper)(nil)

// ClientUnmarshaler defines an interface for unmarshaling ICS02 interfaces.
type ClientUnmarshaler interface {
	MustUnmarshalClientState([]byte) exported.ClientState
	MustUnmarshalConsensusState([]byte) exported.ConsensusState
}

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding client type.
func NewDecodeStore(cdc ClientUnmarshaler, kvA, kvB kv.Pair) (string, bool) {
	switch {
	case bytes.HasPrefix(kvA.Key, host.KeyClientStorePrefix) && bytes.HasSuffix(kvA.Key, host.KeyClientState()):
		clientStateA := cdc.MustUnmarshalClientState(kvA.Value)
		clientStateB := cdc.MustUnmarshalClientState(kvB.Value)
		return fmt.Sprintf("ClientState A: %v\nClientState B: %v", clientStateA, clientStateB), true

	case bytes.HasPrefix(kvA.Key, host.KeyClientStorePrefix) && bytes.HasSuffix(kvA.Key, host.KeyClientType()):
		return fmt.Sprintf("Client type A: %s\nClient type B: %s", string(kvA.Value), string(kvB.Value)), true

	case bytes.HasPrefix(kvA.Key, host.KeyClientStorePrefix) && bytes.Contains(kvA.Key, []byte("consensusState")):
		consensusStateA := cdc.MustUnmarshalConsensusState(kvA.Value)
		consensusStateB := cdc.MustUnmarshalConsensusState(kvB.Value)
		return fmt.Sprintf("ConsensusState A: %v\nConsensusState B: %v", consensusStateA, consensusStateB), true

	default:
		return "", false
	}
}
