package group

import (
	"github.com/cosmos/cosmos-sdk/x/group/internal/orm"
)

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "group"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

func AccountCondition(id uint64) Condition {
	return NewCondition("group", "account", orm.EncodeSequence(id))
}
