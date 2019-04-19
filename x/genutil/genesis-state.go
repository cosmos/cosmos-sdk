package genutil

import "encoding/json"

// State to Unmarshal
type GenesisState struct {
	Accounts []GenesisAccount  `json:"accounts"`
	GenTxs   []json.RawMessage `json:"gentxs"`
}
