package types

import (
	gogoproto "github.com/cosmos/gogoproto/types/any"
)

type Any = gogoproto.Any

var NewAnyWithValue = gogoproto.NewAnyWithCacheWithValue

var UnsafePackAny = gogoproto.UnsafePackAnyWithCache
