package types

import (
	"fmt"

	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// IBC client events
const (
	EventTypeCreateClient       = "create_client"
	EventTypeUpdateClient       = "update_client"
	EventTypeSubmitMisbehaviour = "submit_misbehaviour"

	AttributeKeyClientID = "client_id"
)

// IBC client events vars
var (
	AttributeValueCategory = fmt.Sprintf("%s_%s", ibctypes.ModuleName, SubModuleName)
)
