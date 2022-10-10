package appmodule

import (
	"context"

	"google.golang.org/grpc"

	"cosmossdk.io/depinject"
)

// Handler describes an ABCI app module handler. It can be injected into a
// depinject container as a one-per-module type (in the pointer variant).
type Handler struct {
	// Services are the msg and query services for the module. Msg services
	// must be annotated with the option cosmos.msg.v1.service = true.
	Services []ServiceImpl

	// BeginBlocker doesn't take or return any special arguments as this
	// is the most stable across Tendermint versions and most common need
	// for modules. Special parameters can be injected and/or returned
	// using custom hooks that the app will provide which may vary from
	// one Tendermint release to another.
	BeginBlocker func(context.Context) error

	// EndBlocker doesn't take or return any special arguments as this
	// is the most stable across Tendermint versions and most common need
	// for modules. Special parameters can be injected and/or returned
	// using custom hooks that the app will provide which may vary from
	// one Tendermint release to another.
	EndBlocker func(context.Context) error

	DefaultGenesis  func(GenesisTarget)
	ValidateGenesis func(GenesisSource) error
	InitGenesis     func(context.Context, GenesisSource) error
	ExportGenesis   func(context.Context, GenesisTarget)

	EventListeners []EventListener

	UpgradeHandlers []UpgradeHandler
}

// RegisterService registers a msg or query service. If the cosmos.msg.v1.service
// option is set to true on the service, then it is registered as a msg service,
// otherwise it is registered as a query service.
func (h *Handler) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	h.Services = append(h.Services, ServiceImpl{
		Desc: desc,
		Impl: impl,
	})
}

// ServiceImpl describes a gRPC service implementation to be registered with
// grpc.ServiceRegistrar.
type ServiceImpl struct {
	Desc *grpc.ServiceDesc
	Impl interface{}
}

func (h *Handler) IsOnePerModuleType() {}

var _ depinject.OnePerModuleType = &Handler{}
var _ grpc.ServiceRegistrar = &Handler{}
