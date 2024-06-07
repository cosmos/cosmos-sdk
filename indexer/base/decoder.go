package indexerbase

type DecoderResolver interface {
	Iterate(func(string, ModuleDecoder) error) error
	LookupDecoder(moduleName string) (ModuleDecoder, error)
}

type IndexableModule interface {
	ModuleDecoder() (ModuleDecoder, error)
}

type ModuleDecoder struct {
	Schema Schema

	// KVDecoder is a function that decodes a key-value pair into an EntityUpdate.
	// If modules pass logical updates directly to the engine and don't require logical decoding of raw bytes,
	// then this function should be nil.
	KVDecoder KVDecoder
}

type KVDecoder = func(key, value []byte) (EntityUpdate, bool, error)
