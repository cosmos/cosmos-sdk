package coretesting

import (
	"context"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
)

func SetupTestEnvironment(moduleName string) (context.Context, appmodulev2.Environment) {
	ctx := Context()

	return ctx, appmodulev2.Environment{
		Logger:             nil,
		BranchService:      nil,
		EventService:       EventsService(ctx, moduleName),
		GasService:         nil,
		HeaderService:      MemHeaderService{},
		QueryRouterService: nil,
		MsgRouterService:   nil,
		TransactionService: nil,
		KVStoreService:     KVStoreService(ctx, moduleName),
		MemStoreService:    nil,
	}
}
