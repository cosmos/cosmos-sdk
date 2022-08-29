package appmodule

import (
	"cosmossdk.io/core/blockinfo"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/gas"
	"cosmossdk.io/core/store"
)

type Service interface {
	store.KVStoreService
	store.MemoryStoreService
	store.TransientStoreService
	event.Service
	blockinfo.Service
	gas.Service
}
