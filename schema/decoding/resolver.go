package decoding

import (
	"sort"

	"cosmossdk.io/schema"
)

// DecoderResolver is an interface that allows indexers to discover and use module decoders.
type DecoderResolver interface {
	// IterateAll iterates over all available module decoders.
	IterateAll(func(moduleName string, cdc schema.ModuleCodec) error) error

	// LookupDecoder looks up a specific module decoder.
	LookupDecoder(moduleName string) (decoder schema.ModuleCodec, found bool, err error)
}

// ModuleSetDecoderResolver returns DecoderResolver that will discover modules implementing
// DecodeableModule in the provided module set.
func ModuleSetDecoderResolver(moduleSet map[string]interface{}) DecoderResolver {
	return &moduleSetDecoderResolver{
		moduleSet: moduleSet,
	}
}

type moduleSetDecoderResolver struct {
	moduleSet map[string]interface{}
}

func (a moduleSetDecoderResolver) IterateAll(f func(string, schema.ModuleCodec) error) error {
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
