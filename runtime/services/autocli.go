package services

import (
	"context"

	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"
	protobuf "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/autocli"

	"github.com/cosmos/cosmos-sdk/types/module"
	cosmosmsg "github.com/cosmos/cosmos-sdk/types/msgservice"
)

// AutoCLIQueryService implements the cosmos.autocli.v1.Query service.
type AutoCLIQueryService struct {
	UnimplementedAutoCLIQueryServer

	moduleOptions map[string]*autocli.ModuleOptions
}

// NewAutoCLIQueryService returns a AutoCLIQueryService for the provided modules.
func NewAutoCLIQueryService(appModules map[string]any) *AutoCLIQueryService {
	return &AutoCLIQueryService{
		moduleOptions: ExtractAutoCLIOptions(appModules),
	}
}

// ExtractAutoCLIOptions extracts autocli ModuleOptions from the provided app modules.
//
// Example Usage:
//
//	ExtractAutoCLIOptions(ModuleManager.Modules)
func ExtractAutoCLIOptions(appModules map[string]any) map[string]*autocli.ModuleOptions {
	moduleOptions := map[string]*autocli.ModuleOptions{}
	for modName, mod := range appModules {
		if autoCliMod, ok := mod.(interface {
			AutoCLIOptions() *autocli.ModuleOptions
		}); ok {
			moduleOptions[modName] = autoCliMod.AutoCLIOptions()
			continue
		}

		cfg := &autocliConfigurator{}

		if mod, ok := mod.(module.HasServices); ok {
			mod.RegisterServices(cfg)
		}

		if mod, ok := mod.(appmodule.HasServices); ok {
			err := mod.RegisterServices(cfg)
			if err != nil {
				panic(err)
			}
		}

		if cfg.Error() != nil {
			panic(cfg.Error())
		}

		haveServices := false
		modOptions := &autocli.ModuleOptions{}
		if cfg.msgServer.serviceName != "" {
			haveServices = true
			modOptions.Tx = &autocli.ServiceCommandDescriptor{
				Service: cfg.msgServer.serviceName,
			}
		}

		if cfg.queryServer.serviceName != "" {
			haveServices = true
			modOptions.Query = &autocli.ServiceCommandDescriptor{
				Service: cfg.queryServer.serviceName,
			}
		}

		if haveServices {
			moduleOptions[modName] = modOptions
		}
	}
	return moduleOptions
}

func (a AutoCLIQueryService) AppOptions(_ context.Context, _ *AppOptionsRequest) (*AppOptionsResponse, error) {
	return &AppOptionsResponse{ModuleOptions: a.moduleOptions}, nil
}

var _ AutoCLIQueryServer = &AutoCLIQueryService{}

// autocliConfigurator allows us to call RegisterServices and introspect the services
type autocliConfigurator struct {
	msgServer     autocliServiceRegistrar
	queryServer   autocliServiceRegistrar
	registryCache *protoregistry.Files
	err           error
}

var _ module.Configurator = &autocliConfigurator{}

func (a *autocliConfigurator) MsgServer() gogogrpc.Server { return &a.msgServer }

func (a *autocliConfigurator) QueryServer() gogogrpc.Server { return &a.queryServer }

func (a *autocliConfigurator) RegisterMigration(string, uint64, module.MigrationHandler) error {
	return nil
}

func (a *autocliConfigurator) RegisterService(sd *grpc.ServiceDesc, ss any) {
	if a.registryCache == nil {
		a.registryCache, a.err = proto.MergedRegistry()
	}

	desc, err := a.registryCache.FindDescriptorByName(protoreflect.FullName(sd.ServiceName))
	if err != nil {
		a.err = err
		return
	}

	if protobuf.HasExtension(desc.Options(), cosmosmsg.E_ServiceV2) {
		a.msgServer.RegisterService(sd, ss)
	} else {
		a.queryServer.RegisterService(sd, ss)
	}
}

func (a *autocliConfigurator) Error() error { return nil }

// autocliServiceRegistrar is used to capture the service name for registered services
type autocliServiceRegistrar struct {
	serviceName string
}

func (a *autocliServiceRegistrar) RegisterService(sd *grpc.ServiceDesc, _ any) {
	a.serviceName = sd.ServiceName
}
