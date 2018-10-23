package circuit

import (
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Parameter space name
const DefaultParamspace = "circuit"

// Parameter key
var (
	MsgTypeKey = []byte("msgtype")
	MsgNameKey = []byte("msgname")
)

// Parameter key table
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable(
		MsgTypeKey, bool(false),
		MsgNameKey, bool(false),
	)
}
