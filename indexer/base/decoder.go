package indexerbase

type ModuleDecoder struct {
	Schema    Schema
	KVDecoder KVDecoder
}

type IndexableModule interface {
	ModuleDecoder() (ModuleDecoder, error)
}

type KVDecoder = func(key, value []byte) (EntityUpdate, bool, error)
