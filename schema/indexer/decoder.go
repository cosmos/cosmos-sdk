package indexer

import (
	"context"

	"cosmossdk.io/schema"
	"cosmossdk.io/schema/listener"
)

type DecodingOptions struct {
	DecoderResolver DecoderResolver

	// SyncSource is the source that will be used do initial indexing of modules with pre-existing
	// state. It is optional, but if it is not provided, indexing can only be starting when a node
	// is synced from genesis.
	SyncSource SyncSource
}

func LogicalDecoder(target listener.Listener, opts DecodingOptions) listener.Listener {
	initialize := target.Initialize
	initializeModuleData := target.InitializeModuleData
	onKVPair := target.OnKVPair

	moduleCodecs := map[string]schema.ModuleCodec{}
	target.Initialize = func(ctx context.Context, data listener.InitializationData) (lastBlock int64, err error) {
		if initialize != nil {
			lastBlock, err = initialize(ctx, data)
			if err != nil {
				return
			}
		}

		if opts.DecoderResolver != nil {
			err = opts.DecoderResolver.Iterate(func(moduleName string, codec schema.ModuleCodec) error {
				moduleCodecs[moduleName] = codec
				if initializeModuleData != nil {
					return initializeModuleData(listener.ModuleInitializationData{
						ModuleName: moduleName,
						Schema:     codec.Schema,
					})
				}
				return nil
			})
			if err != nil {
				return
			}
		}

		// TODO: catch-up sync

		return
	}

	target.OnKVPair = func(data listener.KVPairData) error {
		if onKVPair != nil {
			return onKVPair(data)
		}

		if target.OnObjectUpdate != nil {
			codec, ok := moduleCodecs[data.ModuleName]
			if !ok {
				// TODO handle discovering a new module
				return nil
			}

			update, ok, err := codec.KVDecoder(data.Key, data.Value, data.Delete)
			if err != nil {
				return err
			}

			if !ok {
				return nil
			}

			return target.OnObjectUpdate(listener.ObjectUpdateData{
				ModuleName: data.ModuleName,
				Update:     update,
			})
		}

		return nil
	}

	return target
}
