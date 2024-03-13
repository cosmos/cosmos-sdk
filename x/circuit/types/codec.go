package types

import (
	"cosmossdk.io/core/registry"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterInterfaces registers the interfaces types with the interface registry.
func RegisterInterfaces(registry registry.LegacyRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAuthorizeCircuitBreaker{},
		&MsgResetCircuitBreaker{},
		&MsgTripCircuitBreaker{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
