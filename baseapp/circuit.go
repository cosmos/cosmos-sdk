package baseapp

import sdk "github.com/cosmos/cosmos-sdk/types"

// CircuitBreaker is an alias of sdk.CircuitBreaker, kept so code that names
// baseapp.CircuitBreaker still compiles.
//
// Deprecated: use sdk.CircuitBreaker.
type CircuitBreaker = sdk.CircuitBreaker
