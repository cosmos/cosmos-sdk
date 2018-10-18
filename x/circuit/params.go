package circuit

import (
	"github.com/cosmos/cosmos-sdk/x/params"
)

const DefaultParamspace = "circuit"

var (
	MsgTypeKey = []byte("msgtype")
	MsgNameKey = []byte("msgname")
)

func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable(
		MsgTypeKey, bool(false),
		MsgNameKey, bool(false),
	)
}
