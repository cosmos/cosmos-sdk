package types

import (
	"fmt"

	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// IBC client events
const (
	AttributeKeyClientID   = "client_id"
	AttributeKeyClientType = "client_type"
)

// IBC client events vars
var (
	EventTypeCreateClient       = TypeMsgCreateClient
	EventTypeUpdateClient       = TypeMsgUpdateClient
	EventTypeSubmitMisbehaviour = TypeClientMisbehaviour

	AttributeValueCategory = fmt.Sprintf("%s_%s", ibctypes.ModuleName, SubModuleName)
)
