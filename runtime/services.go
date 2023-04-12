package runtime

import (
	"context"

	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/types/module"
)

func (a *App) registerRuntimeServices(cfg module.Configurator) error {
	appv1alpha1.RegisterQueryServer(cfg.QueryServer(), services.NewAppQueryService(a.appConfig))
	autocliv1.RegisterQueryServer(cfg.QueryServer(), services.NewAutoCLIQueryService(a.ModuleManager.Modules))

	reflectionSvc, err := services.NewReflectionService()
	if err != nil {
		return err
	}
	reflectionv1.RegisterReflectionServiceServer(cfg.QueryServer(), reflectionSvc)

	return nil
}

// ======================================================
// ValidatorUpdateService & BlockInfoService
// ======================================================

// ValidatorUpdateService is the extension interface that modules should implement
// if they are conducting validator set updates
type ValidatorUpdateService interface {
	SetValidatorUpdates(context.Context, []abci.ValidatorUpdate)
}

// BlockInfoService is the extension interface that modules should implement
// if they require block information
type BlockInfoService interface {
	GetHeight() int64                // GetHeight returns the height of the block
	Misbehavior() []abci.Misbehavior // Misbehavior returns the misbehavior of the block
	GetHeaderHash() []byte           // GetHeaderHash returns the hash of the block header
	// GetValidatorsHash returns the hash of the validators
	// For Comet, it is the hash of the next validators
	GetValidatorsHash() []byte
	GetProposerAddress() []byte            // GetProposerAddress returns the address of the block proposer
	GetDecidedLastCommit() abci.CommitInfo // GetDecidedLastCommit returns the last commit info
}
