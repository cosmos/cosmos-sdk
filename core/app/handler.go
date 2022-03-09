package app

import (
	"context"

	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/container"
)

func NewHandler() *Handler {
	return &Handler{}
}

// Handler describes an ABCI app handler.
type Handler struct {
	MsgServices   []ServiceImpl
	QueryServices []ServiceImpl

	// BeginBlocker doesn't take or return any special arguments as this
	// is the most stable across Tendermint versions and most common need
	// for modules. Special parameters can be inspected and/or returned
	// using custom hooks that the app will provide which may vary from
	// one Tendermint release to another.
	BeginBlocker func(context.Context) error

	// EndBlocker doesn't take or return any special arguments as this
	// is the most stable across Tendermint versions and most common need
	// for modules. Special parameters can be inspected and/or returned
	// using custom hooks that the app will provide which may vary from
	// one Tendermint release to another.
	EndBlocker func(context.Context) error
}

// RegisterService registers a msg or query service. If the cosmos.msg.v1.service
// option is set true on the service, then it is registered as a msg service,
// otherwise it is registered as a query service.
func (h *Handler) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	//TODO implement me
	panic("implement me")
}

type ServiceImpl struct {
	Desc *grpc.ServiceDesc
	Impl interface{}
}

func (h *Handler) IsOnePerModuleType() {}

var _ container.OnePerModuleType = &Handler{}
var _ grpc.ServiceRegistrar = &Handler{}
