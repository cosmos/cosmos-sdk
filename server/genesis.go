package server

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

var _ sdk.Msg = GenesisMsg{}

// GenesisMsg defines the properties of a genesis message.
type GenesisMsg struct {
	NodeID    string                   `json:"node_id"`
	IP        string                   `json:"ip"`
	Validator tmtypes.GenesisValidator `json:"validator"`
	AppGenTx  json.RawMessage          `json:"app_gen_tx"`
}

//nolint
func (msg GenesisMsg) Type() string             { return "server" }
func (msg GenesisMsg) Name() string             { return "genesis" }
func (msg GenesisMsg) ValidateBasic() sdk.Error { return nil }
func (msg GenesisMsg) GetSignBytes() []byte     { return sdk.MustSortJSON(mustMarshalJSON(msg)) }

//nolint
func (msg GenesisMsg) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.Validator.PubKey.Address())}
}

func mustMarshalJSON(v interface{}) []byte {
	bz, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return bz
}
