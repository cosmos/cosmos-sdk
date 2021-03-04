package types

import (
	"fmt"

	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
)

// IBC client events
const (
	AttributeKeyClientID        = "client_id"
	AttributeKeyClientType      = "client_type"
	AttributeKeyConsensusHeight = "consensus_height"
	AttributeKeyHeader          = "header"
)

// IBC client events vars
var (
	EventTypeCreateClient         = "create_client"
	EventTypeUpdateClient         = "update_client"
	EventTypeUpgradeClient        = "upgrade_client"
	EventTypeSubmitMisbehaviour   = "client_misbehaviour"
	EventTypeUpdateClientProposal = "update_client_proposal"

	AttributeValueCategory = fmt.Sprintf("%s_%s", host.ModuleName, SubModuleName)
)
