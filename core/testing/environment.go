package coretesting

import (
	"context"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	corecontext "cosmossdk.io/core/context"
	corelog "cosmossdk.io/core/log"
	"cosmossdk.io/core/router"
	"cosmossdk.io/core/store"
)

type TestEnvironmentConfig struct {
	ModuleName  string
	Logger      corelog.Logger
	MsgRouter   router.Service
	QueryRouter router.Service
}

type TestEnvironment struct {
	appmodulev2.Environment

	memEventsService MemEventsService
	memHeaderService MemHeaderService
}

func NewTestEnvironment(cfg TestEnvironmentConfig) (TestContext, TestEnvironment) {
	ctx := Context()

	memEventService := EventsService(ctx, cfg.ModuleName)
	memHeaderService := MemHeaderService{}

	env := TestEnvironment{
		Environment: appmodulev2.Environment{
			Logger:             cfg.Logger,
			BranchService:      nil,
			EventService:       memEventService,
			GasService:         MemGasService{},
			HeaderService:      memHeaderService,
			QueryRouterService: cfg.QueryRouter,
			MsgRouterService:   cfg.MsgRouter,
			TransactionService: MemTransactionService{},
			KVStoreService:     KVStoreService(ctx, cfg.ModuleName),
			MemStoreService:    nil,
		},
		memEventsService: memEventService,
		memHeaderService: memHeaderService,
	}

	// set internal context to point to environment
	ctx.Context = context.WithValue(ctx.Context, corecontext.EnvironmentContextKey, env.Environment)
	return ctx, env
}

func (env TestEnvironment) MemEventsService() MemEventsService {
	return env.memEventsService
}

func (env TestEnvironment) KVStoreService() store.KVStoreService {
	return env.Environment.KVStoreService
}

func (env TestEnvironment) HeaderService() MemHeaderService {
	return env.memHeaderService
}
