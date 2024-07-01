package decoding

import (
	"cosmossdk.io/schema"
	"cosmossdk.io/schema/appdata"
)

type Options struct {
	DecoderResolver DecoderResolver

	// SyncSource is the source that will be used do initial indexing of modules with pre-existing
	// state. It is optional, but if it is not provided, indexing can only be starting when a node
	// is synced from genesis.
	SyncSource SyncSource
}

func Middleware(target appdata.Listener, opts Options) (appdata.Listener, error) {
	initializeModuleData := target.InitializeModuleData
	onKVPair := target.OnKVPair

	moduleCodecs := map[string]schema.ModuleCodec{}
	if opts.DecoderResolver != nil {
		err := opts.DecoderResolver.Iterate(func(moduleName string, codec schema.ModuleCodec) error {
			moduleCodecs[moduleName] = codec
			if initializeModuleData != nil {
				return initializeModuleData(appdata.ModuleInitializationData{
					ModuleName: moduleName,
					Schema:     codec.Schema,
				})
			}
			return nil
		})
		if err != nil {
			return appdata.Listener{}, err
		}
	}

	// TODO: catch-up sync

	target.OnKVPair = func(data appdata.KVPairData) error {
		if onKVPair != nil {
			return onKVPair(data)
		}

		if target.OnObjectUpdate != nil {
			for _, kvUpdate := range data.Updates {

				codec, ok := moduleCodecs[kvUpdate.ModuleName]
				if !ok {
					// TODO handle discovering a new module
					return nil
				}

				updates, err := codec.KVDecoder(kvUpdate.Update)
				if err != nil {
					return err
				}

				if !ok {
					return nil
				}

				return target.OnObjectUpdate(appdata.ObjectUpdateData{
					ModuleName: kvUpdate.ModuleName,
					Updates:    updates,
				})
			}
		}

		return nil
	}

	return target, nil
}
