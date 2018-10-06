package params

import (
	"github.com/cosmos/cosmos-sdk/x/params/store"
)

// re-export types from store
type (
	Store         = store.Store
	ReadOnlyStore = store.ReadOnlyStore
	ParamStruct   = store.ParamStruct
	KeyValuePairs = store.KeyValuePairs
)
