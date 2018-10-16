package circuit

import (
	"github.com/cosmos/cosmos-sdk/x/params"
)

const DefaultParamspace = "circuit"

var (
	MsgTypeKey = []byte("msgtype")
	MsgNameKey = []byte("msgname")
)

func ParamTypeTable() params.TypeTable {
	return params.NewTypeTable(
		MsgTypeKey, bool(false),
		MsgNameKey, bool(false),
	)
}
