package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/ibc/common"
)

// IBC connection events
const (
	AttributeKeyConnectionID         = "connection_id"
	AttributeKeyClientID             = "client_id"
	AttributeKeyCounterpartyClientID = "counterparty_client_id"
)

// IBC connection events vars
var (
	EventTypeConnectionOpenInit    = MsgConnectionOpenInit{}.Type()
	EventTypeConnectionOpenTry     = MsgConnectionOpenTry{}.Type()
	EventTypeConnectionOpenAck     = MsgConnectionOpenAck{}.Type()
	EventTypeConnectionOpenConfirm = MsgConnectionOpenConfirm{}.Type()

	AttributeValueCategory = fmt.Sprintf("%s_%s", common.ModuleName, SubModuleName)
)
