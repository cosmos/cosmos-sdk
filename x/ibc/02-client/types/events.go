package types

import (
	"fmt"

	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
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

	AttributeValueCategory = fmt.Sprintf("%s_%s", host.ModuleName, SubModuleName)
)
