package types

import (
	"github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/core/registry"

	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterInterfaces registers the interfaces types with the interface registry.
func RegisterInterfaces(registrar registry.InterfaceRegistrar) {
	registrar.RegisterImplementations((*proto.Message)(nil),
		&MsgAuthorizeCircuitBreaker{},
		&MsgResetCircuitBreaker{},
		&MsgTripCircuitBreaker{},
	)
	msgservice.RegisterMsgServiceDesc(registrar, &_Msg_serviceDesc)
}
