package services

import (
	"context"

	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"
	protobuf "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/autocli"

	"github.com/cosmos/cosmos-sdk/types/module"
	cosmosmsg "github.com/cosmos/cosmos-sdk/types/msgservice"
)

// AutoCLIQueryService implements the cosmos.autocli.v1.Query service.
type AutoCLIQueryService struct {
	autocliv1.UnimplementedQueryServer

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

		// try to auto-discover options based on the last msg and query
		// services registered for the module
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

func (a AutoCLIQueryService) AppOptions(context.Context, *autocliv1.AppOptionsRequest) (*autocliv1.AppOptionsResponse, error) {
	return &autocliv1.AppOptionsResponse{
		ModuleOptions: toProtoModuleOptions(a.moduleOptions),
	}, nil
}

// toProtoModuleOptions converts Go-native autocli options to the pulsar proto
// types required by the gRPC AppOptions endpoint. This is the only place in
// the codebase where that conversion is needed.
func toProtoModuleOptions(opts map[string]*autocli.ModuleOptions) map[string]*autocliv1.ModuleOptions {
	result := make(map[string]*autocliv1.ModuleOptions, len(opts))
	for name, opt := range opts {
		if opt == nil {
			continue
		}
		result[name] = &autocliv1.ModuleOptions{
			Tx:    toProtoServiceDesc(opt.Tx),
			Query: toProtoServiceDesc(opt.Query),
		}
	}
	return result
}

func toProtoServiceDesc(d *autocli.ServiceCommandDescriptor) *autocliv1.ServiceCommandDescriptor {
	if d == nil {
		return nil
	}
	proto := &autocliv1.ServiceCommandDescriptor{
		Service:              d.Service,
		EnhanceCustomCommand: d.EnhanceCustomCommand,
		Short:                d.Short,
	}
	for _, opt := range d.RpcCommandOptions {
		proto.RpcCommandOptions = append(proto.RpcCommandOptions, toProtoRpcOpts(opt))
	}
	if len(d.SubCommands) > 0 {
		proto.SubCommands = make(map[string]*autocliv1.ServiceCommandDescriptor, len(d.SubCommands))
		for k, v := range d.SubCommands {
			proto.SubCommands[k] = toProtoServiceDesc(v)
		}
	}
	return proto
}

func toProtoRpcOpts(o *autocli.RpcCommandOptions) *autocliv1.RpcCommandOptions {
	if o == nil {
		return nil
	}
	proto := &autocliv1.RpcCommandOptions{
		RpcMethod:   o.RpcMethod,
		Use:         o.Use,
		Long:        o.Long,
		Short:       o.Short,
		Example:     o.Example,
		Alias:       o.Alias,
		SuggestFor:  o.SuggestFor,
		Deprecated:  o.Deprecated,
		Version:     o.Version,
		Skip:        o.Skip,
		GovProposal: o.GovProposal,
	}
	for _, p := range o.PositionalArgs {
		proto.PositionalArgs = append(proto.PositionalArgs, &autocliv1.PositionalArgDescriptor{
			ProtoField: p.ProtoField,
			Varargs:    p.Varargs,
			Optional:   p.Optional,
		})
	}
	if len(o.FlagOptions) > 0 {
		proto.FlagOptions = make(map[string]*autocliv1.FlagOptions, len(o.FlagOptions))
		for k, f := range o.FlagOptions {
			if f == nil {
				continue
			}
			proto.FlagOptions[k] = &autocliv1.FlagOptions{
				Name:                f.Name,
				Shorthand:           f.Shorthand,
				Usage:               f.Usage,
				DefaultValue:        f.DefaultValue,
				Deprecated:          f.Deprecated,
				ShorthandDeprecated: f.ShorthandDeprecated,
				Hidden:              f.Hidden,
			}
		}
	}
	return proto
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

var _ autocliv1.QueryServer = &AutoCLIQueryService{}
