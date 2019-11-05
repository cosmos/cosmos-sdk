package types

import (
	"fmt"

	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// IBC client events
const (
	AttributeKeyClientID = "client_id"
)

// IBC client events vars
var (
	EventTypeCreateClient       = MsgCreateClient{}.Type()
	EventTypeUpdateClient       = MsgUpdateClient{}.Type()
	EventTypeSubmitMisbehaviour = MsgSubmitMisbehaviour{}.Type()

	AttributeValueCategory = fmt.Sprintf("%s_%s", ibctypes.ModuleName, SubModuleName)
)
