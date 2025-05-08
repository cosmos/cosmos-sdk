package types

import (
	gogoany "github.com/cosmos/gogoproto/types/any"
)

// Any is an alias for github.com/cosmos/gogoproto/types/any.Any.
type Any = gogoany.Any

// NewAnyWithValue is an alias for github.com/cosmos/gogoproto/types/any.NewAnyWithCacheWithValue.
var NewAnyWithValue = gogoany.NewAnyWithCacheWithValue

// UnsafePackAny is an alias for github.com/cosmos/gogoproto/types/any.UnsafePackAnyWithCache.
var UnsafePackAny = gogoany.UnsafePackAnyWithCache
