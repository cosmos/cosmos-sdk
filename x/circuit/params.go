package circuit

import (
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Parameter space type
const DefaultParamspace = "circuit"

// Parameter key
var (
	MsgRouteKey = []byte("msgroute")
	MsgTypeKey  = []byte("msgtype")
)

// Parameter key table
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable(
		MsgRouteKey, bool(false),
		MsgTypeKey, bool(false),
	)
}
