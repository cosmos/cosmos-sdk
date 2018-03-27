package types

import (
	"encoding/json"
)

// Run only once on chain initialization, should write genesis state to store
// or throw an error if some required information was not provided, in which case
// the application will panic.
type InitGenesis func(ctx Context, data json.RawMessage) error
