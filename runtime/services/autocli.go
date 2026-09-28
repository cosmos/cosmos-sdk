package services

import (
	"context"
	"strings"

	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"
	protobuf "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	cosmosmsg "cosmossdk.io/api/cosmos/msg/v1"
	"cosmossdk.io/core/appmodule"

	"github.com/cosmos/cosmos-sdk/types/module"
)

// AutoCLIQueryService implements the cosmos.autocli.v1.Query service.
type AutoCLIQueryService struct {
	autocliv1.UnimplementedQueryServer

	moduleOptions map[string]*autocliv1.ModuleOptions
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
func ExtractAutoCLIOptions(appModules map[string]any) map[string]*autocliv1.ModuleOptions {
	moduleOptions := map[string]*autocliv1.ModuleOptions{}
	for modName, mod := range appModules {
		if autoCliMod, ok := mod.(interface {
			AutoCLIOptions() *autocliv1.ModuleOptions
		}); ok {
			moduleOptions[modName] = autoCliMod.AutoCLIOptions()
			continue
		}

		cfg := &autocliConfigurator{}

		// try to auto-discover options based on the msg and query services
		// registered for the module
		if mod, ok := mod.(module.HasServices); ok {
			mod.RegisterServices(cfg)
		}

		if mod, ok := mod.(appmodule.HasServices); ok {
			err := mod.RegisterServices(cfg)
			if err != nil {
				panic(err)
			}
		}

		// check for errors in the configurator
		if cfg.Error() != nil {
			panic(cfg.Error())
		}

		modOptions := &autocliv1.ModuleOptions{}
		modOptions.Tx = newServiceCommandDescriptor(cfg.msgServer.serviceNames)
		modOptions.Query = newServiceCommandDescriptor(cfg.queryServer.serviceNames)

		if modOptions.Tx != nil || modOptions.Query != nil {
			moduleOptions[modName] = modOptions
		}
	}
	return moduleOptions
}

func (a AutoCLIQueryService) AppOptions(context.Context, *autocliv1.AppOptionsRequest) (*autocliv1.AppOptionsResponse, error) {
	return &autocliv1.AppOptionsResponse{
		ModuleOptions: a.moduleOptions,
	}, nil
}

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

	if protobuf.HasExtension(desc.Options(), cosmosmsg.E_Service) {
		a.msgServer.RegisterService(sd, ss)
	} else {
		a.queryServer.RegisterService(sd, ss)
	}
}
func (a *autocliConfigurator) Error() error { return nil }

// autocliServiceRegistrar captures the names of every service registered for a
// module. A module may register more than one msg or query service, so all of
// them are retained instead of only the last one.
type autocliServiceRegistrar struct {
	serviceNames []string
}

func (a *autocliServiceRegistrar) RegisterService(sd *grpc.ServiceDesc, _ any) {
	a.serviceNames = append(a.serviceNames, sd.ServiceName)
}

// newServiceCommandDescriptor builds a ServiceCommandDescriptor from the service
// names registered for a single command type (tx or query). It returns nil when
// no service was registered.
//
// The first service is exposed as the primary command. Any additional services
// are attached as sub-commands so they are no longer silently dropped. The
// sub-command name defaults to the lowercased short name of the service, falling
// back to the fully qualified name when that would collide.
func newServiceCommandDescriptor(serviceNames []string) *autocliv1.ServiceCommandDescriptor {
	if len(serviceNames) == 0 {
		return nil
	}

	desc := &autocliv1.ServiceCommandDescriptor{Service: serviceNames[0]}
	if len(serviceNames) == 1 {
		return desc
	}

	desc.SubCommands = make(map[string]*autocliv1.ServiceCommandDescriptor, len(serviceNames)-1)
	for _, name := range serviceNames[1:] {
		key := subCommandKey(name, desc.SubCommands)
		desc.SubCommands[key] = &autocliv1.ServiceCommandDescriptor{Service: name}
	}
	return desc
}

// subCommandKey derives a unique sub-command name for a service. It prefers the
// lowercased short name (the segment after the last dot) and falls back to the
// lowercased fully qualified name when the short name is empty or already taken.
func subCommandKey(serviceName string, taken map[string]*autocliv1.ServiceCommandDescriptor) string {
	key := serviceName
	if i := strings.LastIndexByte(serviceName, '.'); i >= 0 {
		key = serviceName[i+1:]
	}
	key = strings.ToLower(key)

	if _, exists := taken[key]; exists || key == "" {
		key = strings.ToLower(serviceName)
	}
	return key
}

var _ autocliv1.QueryServer = &AutoCLIQueryService{}
