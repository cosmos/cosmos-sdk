package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/ibc/common"
)

// IBC client events
const (
	AttributeKeyClientID   = "client_id"
	AttributeKeyClientType = "client_type"
)

// IBC client events vars
var (
	EventTypeCreateClient       = "create_client"
	EventTypeUpdateClient       = "update_client"
	EventTypeSubmitMisbehaviour = "client_misbehaviour"

	AttributeValueCategory = fmt.Sprintf("%s_%s", common.ModuleName, SubModuleName)
)
