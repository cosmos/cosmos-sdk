package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/sanction/keeper"
)

func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.HasPrefix(kvA.Key, keeper.ParamsPrefix):
			return fmt.Sprintf("%s\n%s", string(kvA.Value), string(kvB.Value))

		case bytes.HasPrefix(kvA.Key, keeper.SanctionedPrefix):
			return fmt.Sprintf("%v\n%v", kvA.Value, kvB.Value)

		case bytes.HasPrefix(kvA.Key, keeper.TemporaryPrefix):
			return fmt.Sprintf("%v\n%v", kvA.Value, kvB.Value)

		case bytes.HasPrefix(kvA.Key, keeper.ProposalIndexPrefix):
			return fmt.Sprintf("%v\n%v", kvA.Value, kvB.Value)

		default:
			panic(fmt.Sprintf("invalid sanction key %X", kvA.Key))
		}
	}
}
