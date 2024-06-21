package decoding

import (
	"sort"

	"cosmossdk.io/schema"
)

type DecoderResolver interface {
	// Iterate iterates over all module decoders which should be initialized at startup.
	Iterate(func(string, schema.ModuleCodec) error) error

	// LookupDecoder allows for resolving decoders dynamically. For instance, some module-like
	// things may come into existence dynamically (like x/accounts or EVM or WASM contracts).
	// The first time the manager sees one of these appearing in KV-store writes, it will
	// lookup a decoder for it and cache it for future use. The manager will also perform
	// a catch-up sync before passing any new writes to ensure that all historical state has
	// been synced if there is any This check will only happen the first time a module is seen
	// by the manager in a given process (a process restart will cause this check to happen again).
	LookupDecoder(moduleName string) (decoder schema.ModuleCodec, found bool, err error)
}

type moduleSetDecoderResolver struct {
	moduleSet map[string]interface{}
}

// ModuleSetDecoderResolver returns DecoderResolver that will discover modules implementing
// DecodeableModule in the provided module set.
func ModuleSetDecoderResolver(moduleSet map[string]interface{}) DecoderResolver {
	return &moduleSetDecoderResolver{
		moduleSet: moduleSet,
	}
}

func (a moduleSetDecoderResolver) Iterate(f func(string, schema.ModuleCodec) error) error {
	keys := make([]string, 0, len(a.moduleSet))
	for k := range a.moduleSet {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		module := a.moduleSet[k]
		dm, ok := module.(schema.HasModuleCodec)
		if ok {
			decoder, err := dm.ModuleCodec()
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

func (a moduleSetDecoderResolver) LookupDecoder(moduleName string) (schema.ModuleCodec, bool, error) {
	mod, ok := a.moduleSet[moduleName]
	if !ok {
		return schema.ModuleCodec{}, false, nil
	}

	dm, ok := mod.(schema.HasModuleCodec)
	if !ok {
		return schema.ModuleCodec{}, false, nil
	}

	decoder, err := dm.ModuleCodec()
	return decoder, true, err
}
