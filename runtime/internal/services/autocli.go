package services

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/core/appmodule"
)

type autocliService struct {
	autocliv1.UnimplementedRemoteInfoServiceServer

	moduleOptions map[string]*autocliv1.ModuleOptions
}

func newAutocliService(appModules map[string]appmodule.AppModule) *autocliService {
	moduleOptions := map[string]*autocliv1.ModuleOptions{}
	for modName, mod := range appModules {
		if autoCliMod, ok := mod.(interface {
			AutoCLIOptions() *autocliv1.ModuleOptions
		}); ok {
			moduleOptions[modName] = autoCliMod.AutoCLIOptions()
		}
	}
	return &autocliService{
		moduleOptions: moduleOptions,
	}
}

func (a autocliService) AppOptions(context.Context, *autocliv1.AppOptionsRequest) (*autocliv1.AppOptionsResponse, error) {
	return &autocliv1.AppOptionsResponse{
		ModuleOptions: a.moduleOptions,
	}, nil
}

var _ autocliv1.RemoteInfoServiceServer = &autocliService{}
