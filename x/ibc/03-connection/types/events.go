package types

import (
	"fmt"

	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// IBC connection events
const (
	EventTypeConnectionOpenInit    = "connection_open_init"
	EventTypeConnectionOpenTry     = "connection_open_try"
	EventTypeConnectionOpenAck     = "connection_open_ack"
	EventTypeConnectionOpenConfirm = "connection_open_confirm"

	AttributeKeyConnectionID         = "connection_id"
	AttributeKeyCounterpartyClientID = "counterparty_client_id"
)

// IBC connection events vars
var (
	AttributeValueCategory = fmt.Sprintf("%s_%s", ibctypes.ModuleName, SubModuleName)
)
