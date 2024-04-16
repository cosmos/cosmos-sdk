package appmodule

import (
	"cosmossdk.io/core/branch"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/gas"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/router"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
)

// Environment is used to get all services to their respective module
type Environment struct {
	Logger log.Logger

	BranchService      branch.Service
	EventService       event.Service
	GasService         gas.Service
	HeaderService      header.Service
	RouterService      router.Service
	TransactionService transaction.Service

	KVStoreService  store.KVStoreService
	MemStoreService store.MemoryStoreService
}
