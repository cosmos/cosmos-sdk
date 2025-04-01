package types

import (
	gogoany "github.com/cosmos/gogoproto/types/any"
)

type Any = gogoany.Any

var NewAnyWithValue = gogoany.NewAnyWithCacheWithValue

var UnsafePackAny = gogoany.UnsafePackAnyWithCache
