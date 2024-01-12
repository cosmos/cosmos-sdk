package appmodule

import (
	"cosmossdk.io/core/branch"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/gas"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/store"
)

// Environment is used to get all services to their respective module
type Environment struct {
	BranchService   branch.Service
	EventService    event.Service
	GasService      gas.Service
	HeaderService   header.Service
	KvStoreService  map[string]store.KVStoreService
	MemStoreService map[string]store.MemoryStoreService
}
