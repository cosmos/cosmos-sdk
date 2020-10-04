package types

import (
	"fmt"

	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
)

// IBC connection events
const (
	AttributeKeyConnectionID             = "connection_id"
	AttributeKeyClientID                 = "client_id"
	AttributeKeyCounterpartyClientID     = "counterparty_client_id"
	AttributeKeyCounterpartyConnectionID = "counterparty_connection_id"
)

// IBC connection events vars
var (
	EventTypeConnectionOpenInit    = MsgConnectionOpenInit{}.Type()
	EventTypeConnectionOpenTry     = MsgConnectionOpenTry{}.Type()
	EventTypeConnectionOpenAck     = MsgConnectionOpenAck{}.Type()
	EventTypeConnectionOpenConfirm = MsgConnectionOpenConfirm{}.Type()

	AttributeValueCategory = fmt.Sprintf("%s_%s", host.ModuleName, SubModuleName)
)
