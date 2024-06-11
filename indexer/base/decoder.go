package indexerbase

import "sort"

type DecoderResolver interface {
	// Iterate iterates over all module decoders which should be initialized at startup.
	Iterate(func(string, ModuleDecoder) error) error

	// LookupDecoder allows for resolving decoders dynamically. For instance, some module-like
	// things may come into existence dynamically (like x/accounts or EVM or WASM contracts).
	// The first time the manager sees one of these appearing in KV-store writes, it will
	// lookup a decoder for it and cache it for future use. This check will only happen the first
	// time a module is seen. In order to start decoding an existing module, the indexing manager
	// needs to be restarted, usually with a node restart and indexer catch-up needs to be run.
	LookupDecoder(moduleName string) (decoder ModuleDecoder, found bool, err error)
}

// DecodableModule is an interface that modules can implement to provide a ModuleDecoder.
// Usually these modules would also implement appmodule.AppModule, but that is not included
// to leave this package free of any dependencies.
type DecodableModule interface {

	// ModuleDecoder returns a ModuleDecoder for the module.
	ModuleDecoder() (ModuleDecoder, error)
}

// ModuleDecoder is a struct that contains the schema and a KVDecoder for a module.
type ModuleDecoder struct {
	// Schema is the schema for the module.
	Schema ModuleSchema

	// KVDecoder is a function that decodes a key-value pair into an EntityUpdate.
	// If modules pass logical updates directly to the engine and don't require logical decoding of raw bytes,
	// then this function should be nil.
	KVDecoder KVDecoder
}

// KVDecoder is a function that decodes a key-value pair into an EntityUpdate.
// If the KV-pair doesn't represent an entity update, the function should return false
// as the second return value. Error should only be non-nil when the decoder expected
// to parse a valid update and was unable to.
type KVDecoder = func(key, value []byte) (EntityUpdate, bool, error)

type appModuleDecoderResolver[ModuleT any] struct {
	moduleSet map[string]ModuleT
}

// NewAppModuleDecoderResolver returns DecoderResolver that will discover modules implementing
// DecodeableModule in the provided module set.
func NewAppModuleDecoderResolver[ModuleT any](moduleSet map[string]ModuleT) DecoderResolver {
	return &appModuleDecoderResolver[ModuleT]{
		moduleSet: moduleSet,
	}
}

func (a appModuleDecoderResolver[ModuleT]) Iterate(f func(string, ModuleDecoder) error) error {
	keys := make([]string, 0, len(a.moduleSet))
	for k := range a.moduleSet {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		module := a.moduleSet[k]
		dm, ok := any(module).(DecodableModule)
		if ok {
			decoder, err := dm.ModuleDecoder()
			if err != nil {
				return err
			}
			err = f(k, decoder)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (a appModuleDecoderResolver[ModuleT]) LookupDecoder(moduleName string) (ModuleDecoder, bool, error) {
	mod, ok := a.moduleSet[moduleName]
	if !ok {
		return ModuleDecoder{}, false, nil
	}

	dm, ok := any(mod).(DecodableModule)
	if !ok {
		return ModuleDecoder{}, false, nil
	}

	decoder, err := dm.ModuleDecoder()
	return decoder, true, err
}
