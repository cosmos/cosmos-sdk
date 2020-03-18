package capability

import (
	"github.com/cosmos/cosmos-sdk/x/capability/keeper"
	"github.com/cosmos/cosmos-sdk/x/capability/types"
)

// nolint

var (
	NewCapabilityKey = types.NewCapabilityKey
	NewKeeper        = keeper.NewKeeper
)

type (
	Capability    = types.Capability
	CapabilityKey = types.CapabilityKey
)
