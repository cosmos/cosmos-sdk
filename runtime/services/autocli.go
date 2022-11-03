package services

import (
	"context"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"github.com/cosmos/cosmos-sdk/types/module"
)

// AutoCLIQueryService implements the cosmos.autocli.v1.Query service.
type AutoCLIQueryService struct {
	autocliv1.UnimplementedQueryServer

	moduleOptions map[string]*autocliv1.ModuleOptions
}

func NewAutoCLIQueryService(appModules map[string]module.AppModule) *AutoCLIQueryService {
	moduleOptions := map[string]*autocliv1.ModuleOptions{}
	for modName, mod := range appModules {
		if autoCliMod, ok := mod.(interface {
			AutoCLIOptions() *autocliv1.ModuleOptions
		}); ok {
			moduleOptions[modName] = autoCliMod.AutoCLIOptions()
		}
	}
	return &AutoCLIQueryService{
		moduleOptions: moduleOptions,
	}
}

func (a AutoCLIQueryService) AppOptions(context.Context, *autocliv1.AppOptionsRequest) (*autocliv1.AppOptionsResponse, error) {
	return &autocliv1.AppOptionsResponse{
		ModuleOptions: a.moduleOptions,
	}, nil
}

var _ autocliv1.QueryServer = &AutoCLIQueryService{}
