package indexing

import (
	"cosmossdk.io/schema/appdata"
	"cosmossdk.io/schema/decoding"
)

type ManagerOptions struct {
	Options    map[string]interface{}
	Resolver   decoding.DecoderResolver
	SyncSource decoding.SyncSource
}

func NewManager(opts ManagerOptions) appdata.Listener {
	panic("TODO")
}
