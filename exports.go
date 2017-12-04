package sdk

import (
	"github.com/cosmos/cosmos-sdk/store"
	types "github.com/cosmos/cosmos-sdk/types"
)

type (
	// Type aliases for the cosmos-sdk/types module.  We keep all of them in
	// types/* but they are all meant to be imported as
	// "github.com/cosmos/cosmos-sdk".  So, add all of them.
	Handler   = types.Handler
	Context   = types.Context
	Decorator = types.Decorator

	// Type aliases for other modules.
	MultiStore = store.MultiStore modules.
)
