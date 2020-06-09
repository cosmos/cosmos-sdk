package simulation

import (
	"bytes"
	"fmt"

	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding client type.
func NewDecodeStore(cdc codec.Marshaler) func(kvA, kvB tmkv.Pair) string {
	return func(kvA, kvB tmkv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key, host.KeyClientStorePrefix) && bytes.HasSuffix(kvA.Key, host.KeyClientState()):
			var clientStateA, clientStateB exported.ClientState
			cdc.MustUnmarshalBinaryBare(kvA.Value, &clientStateA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &clientStateB)
			return fmt.Sprintf("ClientState A: %v\nClientState B:%v", clientStateA, clientStateB)

		case bytes.Equal(kvA.Key, host.KeyClientStorePrefix) && bytes.HasSuffix(kvA.Key, host.KeyClientType()):
			return fmt.Sprintf("Client type A: %v\nClient type B:%v", string(kvA.Value), string(kvB.Value))

		case bytes.Equal(kvA.Key, host.KeyClientStorePrefix) && bytes.Contains(kvA.Key, []byte("consensusState")):
			var consensusStateA, consensusStateB exported.ConsensusState
			cdc.MustUnmarshalBinaryBare(kvA.Value, &consensusStateA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &consensusStateB)
			return fmt.Sprintf("ConsensusState A: %v\nConsensusState B:%v", consensusStateA, consensusStateB)

		default:
			panic(fmt.Sprintf("invalid %s %s key prefix: %s", host.ModuleName, types.SubModuleName, string(kvA.Key)))
		}
	}
}
