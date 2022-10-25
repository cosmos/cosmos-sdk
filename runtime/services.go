package runtime

import (
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"

	"github.com/cosmos/cosmos-sdk/runtime/internal/services"
)

func (a *App) registerRuntimeServices() error {
	appConfigSvc, err := services.NewAppConfigService(a.appConfig)
	if err != nil {
		return err
	}
	appv1alpha1.RegisterQueryServer(a.GRPCQueryRouter(), appConfigSvc)

	autocliv1.RegisterQueryServer(a.GRPCQueryRouter(), services.NewAutocliService(a.appModules))

	reflectionSvc, err := services.NewReflectionService()
	if err != nil {
		return err
	}
	reflectionv1.RegisterReflectionServiceServer(a.GRPCQueryRouter(), reflectionSvc)

	return nil
}
