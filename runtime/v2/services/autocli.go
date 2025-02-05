package services

import (
	"context"
	"fmt"

	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"
	protobuf "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	cosmosmsg "cosmossdk.io/api/cosmos/msg/v1"
	"cosmossdk.io/core/appmodule"
)

// AutoCLIQueryService implements the cosmos.autocli.v1.Query service.
type AutoCLIQueryService struct {
	autocliv1.UnimplementedQueryServer

	ModuleOptions map[string]*autocliv1.ModuleOptions
}

// ExtractAutoCLIOptions extracts autocli ModuleOptions from the provided app modules.
// Example Usage: ExtractAutoCLIOptions(ModuleManager.Modules)
// Note, runtimev2/services.ExtractAutoCLIOptions differs from runtimev1/services.ExtractAutoCLIOptions as
// it supports only modules implementing fully the core appmodule interface.
func ExtractAutoCLIOptions(appModules map[string]appmodule.AppModule) (map[string]*autocliv1.ModuleOptions, error) {
	moduleOptions := map[string]*autocliv1.ModuleOptions{}
	for modName, mod := range appModules {
		if autoCliMod, ok := mod.(interface {
			AutoCLIOptions() *autocliv1.ModuleOptions
		}); ok {
			moduleOptions[modName] = autoCliMod.AutoCLIOptions()
			continue
		}

		autoCliRegistrar := &autocliRegistrar{}

		// try to auto-discover options based on the last msg and query
		// services registered for the module
		// this supports the legacy core/appmodule v1 service registration
		if mod, ok := mod.(interface {
			RegisterServices(grpc.ServiceRegistrar) error
		}); ok {
			err := mod.RegisterServices(autoCliRegistrar)
			if err != nil {
				return nil, err
			}
		}

		// check for errors in the registrar
		if err := autoCliRegistrar.Error(); err != nil {
			return nil, err
		}

		haveServices := false
		modOptions := &autocliv1.ModuleOptions{}
		if autoCliRegistrar.msgServer.serviceName != "" {
			haveServices = true
			modOptions.Tx = &autocliv1.ServiceCommandDescriptor{
				Service: autoCliRegistrar.msgServer.serviceName,
			}
		}

		if autoCliRegistrar.queryServer.serviceName != "" {
			haveServices = true
			modOptions.Query = &autocliv1.ServiceCommandDescriptor{
				Service: autoCliRegistrar.queryServer.serviceName,
			}
		}

		if haveServices {
			moduleOptions[modName] = modOptions
		}
	}
	return moduleOptions, nil
}

func (a AutoCLIQueryService) AppOptions(context.Context, *autocliv1.AppOptionsRequest) (*autocliv1.AppOptionsResponse, error) {
	return &autocliv1.AppOptionsResponse{
		ModuleOptions: a.ModuleOptions,
	}, nil
}

var _ grpc.ServiceRegistrar = (*autocliRegistrar)(nil)

// autocliRegistrar allows us to call RegisterServices and introspect the services
type autocliRegistrar struct {
	msgServer     autocliServiceRegistrar
	queryServer   autocliServiceRegistrar
	registryCache *protoregistry.Files
	err           error
}

func (a *autocliRegistrar) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	if a.registryCache == nil {
		a.registryCache, a.err = proto.MergedRegistry()
		if a.err != nil {
			a.err = fmt.Errorf("failed to build registry cache: %w", a.err)
			return
		}
	}

	fullName := protoreflect.FullName(sd.ServiceName)
	desc, err := a.registryCache.FindDescriptorByName(fullName)
	if err != nil {
		a.err = fmt.Errorf("failed to find descriptor for %q: %w", fullName, err)
		return
	}

	if protobuf.HasExtension(desc.Options(), cosmosmsg.E_Service) {
		a.msgServer.RegisterService(sd, ss)
	} else {
		a.queryServer.RegisterService(sd, ss)
	}
}

func (a *autocliRegistrar) Error() error {
	return a.err
}

// autocliServiceRegistrar is used to capture the service name for registered services
type autocliServiceRegistrar struct {
	serviceName string
}

func (a *autocliServiceRegistrar) RegisterService(sd *grpc.ServiceDesc, _ interface{}) {
	a.serviceName = sd.ServiceName
}

var _ autocliv1.QueryServer = &AutoCLIQueryService{}
