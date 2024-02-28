package simapp

import (
	"encoding/json"
)

// GenesisState of the blockchain is represented here as a map of raw json
// messages key'd by a identifier string.
// The identifier is used to determine which module genesis information belongs
// to so it may be appropriately routed during init chain.
// Within this application default genesis information is retrieved from
// the module manager which populates json from each module
// object provided to it during init.
type GenesisState map[string]json.RawMessage
