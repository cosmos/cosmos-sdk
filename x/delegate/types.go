package delegate

import (
	cosmos "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/types"
)

type Action interface {
	cosmos.Msg
	Actor() cosmos.AccAddress
	RequiredCapabilities() []Capability
}

type Capability interface {
	// Every type of capability should be have a system wide unique key
	CapabilityKey() string
	// Accept determines whether this grant allows the provided action, and if
	// so provides an upgraded capability grant
	Accept(action Action, block abci.Header) (allow bool, updated Capability, delete bool)
}

type ActorCapability struct {
	Capability Capability
	Actor      cosmos.AccAddress
}

type Keeper interface {
	// Store capabilities under the key actor-id/capability-id
	// Grant stores a root flag, and delegate
	//GrantRootCapability(ctx cosmos.Context, actor cosmos.AccAddress, capability Capability)
	//RevokeRootCapability(ctx cosmos.Context, actor cosmos.AccAddress, capability Capability)
	Delegate(ctx cosmos.Context, grantor cosmos.AccAddress, grantee cosmos.AccAddress, capability ActorCapability) bool
	Undelegate(ctx cosmos.Context, grantor cosmos.AccAddress, grantee cosmos.AccAddress, capability ActorCapability)
	HasCapability(ctx cosmos.Context, actor cosmos.AccAddress, capability ActorCapability) bool
}

type Dispatcher interface {
	DispatchAction(ctx cosmos.Context, actor cosmos.AccAddress, action Action) cosmos.Result
}
