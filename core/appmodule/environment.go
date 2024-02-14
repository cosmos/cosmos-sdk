package appmodule

import (
	"cosmossdk.io/core/branch"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/gas"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
)

// Environment is used to get all services to their respective module
type Environment struct {
	BranchService   branch.Service
	EventService    event.Service
	GasService      gas.Service
	HeaderService   header.Service
	KVStoreService  store.KVStoreService
	MemStoreService store.MemoryStoreService
	Logger          log.Logger
}
