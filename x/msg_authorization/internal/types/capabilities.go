package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

type Capability interface {
	MsgType() sdk.Msg
	Accept(msg sdk.Msg, block abci.Header) (allow bool, updated Capability, delete bool)
}

type CapabilityGrant struct {
	Capability Capability

	Expiration time.Time
}

// GenericCapability grants the permission to execute any transaction of the provided
// sdk.Msg type without restrictions
type GenericCapability struct {
	// MsgType is the type of Msg this capability grant allows
	MsgType sdk.Msg
}
