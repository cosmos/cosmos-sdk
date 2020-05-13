package capability

import (
	"github.com/cosmos/cosmos-sdk/x/capability/keeper"
	"github.com/cosmos/cosmos-sdk/x/capability/types"
)

// DONTCOVER

const (
	ModuleName  = types.ModuleName
	StoreKey    = types.StoreKey
	MemStoreKey = types.MemStoreKey
)

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
	NewOwner                 = types.NewOwner
	NewCapabilityOwners      = types.NewCapabilityOwners
)

type (
	Keeper        = keeper.Keeper
	ScopedKeeper  = keeper.ScopedKeeper
	Capability    = types.Capability
	Owners        = types.CapabilityOwners
	GenesisState  = types.GenesisState
	GenesisOwners = types.GenesisOwners
)
