package coretesting

import (
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/router"
)

func SetupTestEnvironment(moduleName string, msgRouter, queryRouter router.Service) (TestContext, appmodulev2.Environment) {
	ctx := Context()

	return ctx, appmodulev2.Environment{
		Logger:             nil,
		BranchService:      nil,
		EventService:       EventsService(ctx, moduleName),
		GasService:         MemGasService{},
		HeaderService:      MemHeaderService{},
		QueryRouterService: queryRouter,
		MsgRouterService:   msgRouter,
		TransactionService: MemTransactionService{},
		KVStoreService:     KVStoreService(ctx, moduleName),
		MemStoreService:    nil,
	}
}
