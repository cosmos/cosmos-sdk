package capability

import (
	"github.com/cosmos/cosmos-sdk/x/capability/keeper"
	"github.com/cosmos/cosmos-sdk/x/capability/types"
)

// nolint
// DONTCOVER

const (
	ModuleName  = types.ModuleName
	StoreKey    = types.StoreKey
	MemStoreKey = types.MemStoreKey
)

var (
	NewKeeper                   = keeper.NewKeeper
	NewCapabilityKey            = types.NewCapabilityKey
	RevCapabilityKey            = types.RevCapabilityKey
	FwdCapabilityKey            = types.FwdCapabilityKey
	KeyIndex                    = types.KeyIndex
	KeyPrefixIndexCapability    = types.KeyPrefixIndexCapability
	ErrCapabilityTaken          = types.ErrCapabilityTaken
	ErrOwnerClaimed             = types.ErrOwnerClaimed
	RegisterCodec               = types.RegisterCodec
	RegisterCapabilityTypeCodec = types.RegisterCapabilityTypeCodec
	ModuleCdc                   = types.ModuleCdc
	NewOwner                    = types.NewOwner
	NewCapabilityOwners         = types.NewCapabilityOwners
)

type (
	Capability       = types.Capability
	CapabilityKey    = types.CapabilityKey
	CapabilityOwners = types.CapabilityOwners
)
