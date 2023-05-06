package simulation

import (
	"bytes"
	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	"fmt"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

func NewStoreDecoderFuncFromCollectionsSchema(schema collections.Schema) func(kvA, kvB kv.Pair) string {
	colls := schema.ListCollections()
	prefixes := make([][]byte, len(colls))
	valueCodecs := make([]collcodec.UntypedValueCodec, len(colls))
	for i, coll := range colls {
		prefixes[i] = coll.GetPrefix()
		valueCodecs[i] = coll.ValueCodec()
	}

	return func(kvA, kvB kv.Pair) string {
		for i, prefix := range prefixes {
			if bytes.HasPrefix(kvA.Key, prefix) {
				vc := valueCodecs[i]
				// unmarshal kvA.Value to the corresponding type
				vA, err := vc.Decode(kvA.Value)
				if err != nil {
					panic(err)
				}
				// unmarshal kvB.Value to the corresponding type
				vB, err := vc.Decode(kvB.Value)
				if err != nil {
					panic(err)
				}
				vAString, err := vc.Stringify(vA)
				if err != nil {
					panic(err)
				}
				vBString, err := vc.Stringify(vB)
				if err != nil {
					panic(err)
				}
				return vAString + "\n" + vBString
			}
		}
		panic(fmt.Errorf("unexpected key %X (%s)", kvA.Key, kvA.Key))
	}
}
