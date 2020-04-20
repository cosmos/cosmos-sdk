package simulation

import (
	"bytes"
	"fmt"

	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/x/evidence/types"
)

// DONTCOVER

// DecodeStore unmarshals the KVPair's Value to the corresponding evidence type
func DecodeStore(cdc types.Codec, kvA, kvB tmkv.Pair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], types.KeyPrefixEvidence):
		evidenceA, err := cdc.UnmarshalEvidence(kvA.Value)
		if err != nil {
			panic(fmt.Sprintf("cannot unmarshal evidence: %s", err.Error()))
		}
		evidenceB, err := cdc.UnmarshalEvidence(kvB.Value)
		if err != nil {
			panic(fmt.Sprintf("cannot unmarshal evidence: %s", err.Error()))
		}
		return fmt.Sprintf("%v\n%v", evidenceA, evidenceB)

	default:
		panic(fmt.Sprintf("invalid %s key prefix %X", types.ModuleName, kvA.Key[:1]))
	}
}
