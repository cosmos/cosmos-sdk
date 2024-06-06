package indexerbase

type IndexableModule interface {
	ModuleDecoder() (ModuleDecoder, error)
}

type ModuleDecoder struct {
	Schema    Schema
	KVDecoder KVDecoder
}

type KVDecoder = func(key, value []byte) (EntityUpdate, bool, error)
