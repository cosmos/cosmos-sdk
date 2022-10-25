package services

import (
	"context"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/core/appmodule"
)

type AutocliService struct {
	autocliv1.UnimplementedQueryServer

	moduleOptions map[string]*autocliv1.ModuleOptions
}

func NewAutocliService(appModules map[string]appmodule.AppModule) *AutocliService {
	moduleOptions := map[string]*autocliv1.ModuleOptions{}
	for modName, mod := range appModules {
		if autoCliMod, ok := mod.(interface {
			AutoCLIOptions() *autocliv1.ModuleOptions
		}); ok {
			moduleOptions[modName] = autoCliMod.AutoCLIOptions()
		}
	}
	return &AutocliService{
		moduleOptions: moduleOptions,
	}
}

func (a AutocliService) AppOptions(context.Context, *autocliv1.AppOptionsRequest) (*autocliv1.AppOptionsResponse, error) {
	return &autocliv1.AppOptionsResponse{
		ModuleOptions: a.moduleOptions,
	}, nil
}

var _ autocliv1.QueryServer = &AutocliService{}
