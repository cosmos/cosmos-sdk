package capability

import (
	"github.com/cosmos/cosmos-sdk/x/capability/keeper"
	"github.com/cosmos/cosmos-sdk/x/capability/types"
)

// DONTCOVER

// nolint
const (
	ModuleName  = types.ModuleName
	StoreKey    = types.StoreKey
	MemStoreKey = types.MemStoreKey
)

// nolint
var (
	NewKeeper                = keeper.NewKeeper
	NewCapability            = types.NewCapability
	RevCapabilityKey         = types.RevCapabilityKey
	FwdCapabilityKey         = types.FwdCapabilityKey
	KeyIndex                 = types.KeyIndex
	KeyPrefixIndexCapability = types.KeyPrefixIndexCapability
	ErrCapabilityTaken       = types.ErrCapabilityTaken
	ErrOwnerClaimed          = types.ErrOwnerClaimed
	ErrCapabilityNotOwned    = types.ErrCapabilityNotOwned
	RegisterCodec            = types.RegisterCodec
	ModuleCdc                = types.ModuleCdc
	NewOwner                 = types.NewOwner
	NewCapabilityOwners      = types.NewCapabilityOwners
	NewCapabilityStore       = types.NewCapabilityStore
)

// nolint
type (
	Keeper           = keeper.Keeper
	ScopedKeeper     = keeper.ScopedKeeper
	Capability       = types.Capability
	CapabilityOwners = types.CapabilityOwners
	CapabilityStore  = types.CapabilityStore
)
