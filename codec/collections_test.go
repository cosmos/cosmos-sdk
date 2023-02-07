package codec

import (
	"testing"

	"cosmossdk.io/collections/colltest"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/gogoproto/types"
)

func TestCollectionsCorrectness(t *testing.T) {
	cdc := NewProtoCodec(codectypes.NewInterfaceRegistry())
	t.Run("CollValue", func(t *testing.T) {
		colltest.TestValueCodec(t, CollValue[types.UInt64Value](cdc), types.UInt64Value{
			Value: 500,
		})
	})
}
