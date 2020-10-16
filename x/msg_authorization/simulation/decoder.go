package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/types"
)

// NewDecodeStore returns a decoder function closure that umarshals the KVPair's
// Value to the corresponding msg_authorization type.
func NewDecodeStore(cdc codec.Marshaler) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.GrantKey):
			var grantA, grantB types.AuthorizationGrant
			cdc.MustUnmarshalBinaryBare(kvA.Value, &grantA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &grantB)
			fmt.Println(grantA)
			return fmt.Sprintf("%v\n%v", grantA, grantB)
		default:
			panic(fmt.Sprintf("invalid msg_authorization key %X", kvA.Key))
		}
	}
}
