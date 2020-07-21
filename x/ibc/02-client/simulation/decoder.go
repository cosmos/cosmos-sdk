package simulation

import (
	"bytes"
	"fmt"

	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding client type.
func NewDecodeStore(cdc *codec.Codec, kvA, kvB tmkv.Pair) (string, bool) {
	switch {
	case bytes.HasPrefix(kvA.Key, host.KeyClientStorePrefix) && bytes.HasSuffix(kvA.Key, host.KeyClientState()):
		var clientStateA, clientStateB exported.ClientState
		cdc.MustUnmarshalBinaryBare(kvA.Value, &clientStateA)
		cdc.MustUnmarshalBinaryBare(kvB.Value, &clientStateB)
		return fmt.Sprintf("ClientState A: %v\nClientState B: %v", clientStateA, clientStateB), true

	case bytes.HasPrefix(kvA.Key, host.KeyClientStorePrefix) && bytes.HasSuffix(kvA.Key, host.KeyClientType()):
		return fmt.Sprintf("Client type A: %s\nClient type B: %s", string(kvA.Value), string(kvB.Value)), true

	case bytes.HasPrefix(kvA.Key, host.KeyClientStorePrefix) && bytes.Contains(kvA.Key, []byte("consensusState")):
		var consensusStateA, consensusStateB exported.ConsensusState
		cdc.MustUnmarshalBinaryBare(kvA.Value, &consensusStateA)
		cdc.MustUnmarshalBinaryBare(kvB.Value, &consensusStateB)
		return fmt.Sprintf("ConsensusState A: %v\nConsensusState B: %v", consensusStateA, consensusStateB), true

	default:
		return "", false
	}
}
