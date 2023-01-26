package orm

import (
	"context"
	"fmt"

	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	modulev1alpha1 "cosmossdk.io/api/cosmos/orm/module/v1alpha1"
	ormv1alpha1 "cosmossdk.io/api/cosmos/orm/v1alpha1"
	"cosmossdk.io/depinject"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/orm/model/ormdb"
	"github.com/cosmos/cosmos-sdk/orm/model/ormtable"
)

func init() {
	appmodule.Register(&modulev1alpha1.Module{},
		appmodule.Provide(ProvideModuleDB),
	)
}

type ModuleDBInputs struct {
	AppConfig             *appv1alpha1.Config
	KVStoreService        store.KVStoreService
	MemStoreService       store.MemoryStoreService    `optional:"true"`
	TransientStoreService store.TransientStoreService `optional:"true"`
	TypeResolver          ormtable.TypeResolver       `optional:"true"`
	FileResolver          protodesc.Resolver          `optional:"true"`
}

func ProvideModuleDB(moduleKey depinject.ModuleKey, inputs ModuleDBInputs) (ormdb.ModuleDB, error) {
	for _, module := range inputs.AppConfig.Modules {
		if module.Name == moduleKey.Name() {
			typeResolver := inputs.TypeResolver
			if typeResolver == nil {
				typeResolver = protoregistry.GlobalTypes
			}

			modTyp, err := typeResolver.FindMessageByURL(module.Config.TypeUrl)
			if err != nil {
				return nil, err
			}

			modSchema := proto.GetExtension(modTyp.Descriptor().Options(), ormv1alpha1.E_ModuleSchema).(*ormv1alpha1.ModuleSchemaDescriptor)
			if modSchema == nil {
				return nil, fmt.Errorf("no schema for module %s", moduleKey.Name())
			}

			return ormdb.NewModuleDB(modSchema, ormdb.ModuleDBOptions{
				TypeResolver: inputs.TypeResolver,
				FileResolver: inputs.FileResolver,
				GetBackendResolver: func(storageType ormv1alpha1.StorageType) (ormtable.BackendResolver, error) {
					switch storageType {
					case ormv1alpha1.StorageType_STORAGE_TYPE_DEFAULT_UNSPECIFIED:
						return func(ctx context.Context) (ormtable.ReadBackend, error) {
							kvStore := inputs.KVStoreService.OpenKVStore(ctx)
							return ormtable.NewBackend(ormtable.BackendOptions{
								CommitmentStore: kvStore,
								IndexStore:      kvStore,
								ValidateHooks:   nil,
								WriteHooks:      nil,
							}), nil
						}, nil
					case ormv1alpha1.StorageType_STORAGE_TYPE_MEMORY:
						return func(ctx context.Context) (ormtable.ReadBackend, error) {
							if inputs.MemStoreService == nil {
								return nil, fmt.Errorf("unsupported backend type %s", storageType)
							}

							kvStore := inputs.MemStoreService.OpenMemoryStore(ctx)
							return ormtable.NewBackend(ormtable.BackendOptions{
								CommitmentStore: kvStore,
								IndexStore:      kvStore,
								ValidateHooks:   nil,
								WriteHooks:      nil,
							}), nil
						}, nil
					case ormv1alpha1.StorageType_STORAGE_TYPE_TRANSIENT:
						return func(ctx context.Context) (ormtable.ReadBackend, error) {
							if inputs.TransientStoreService == nil {
								return nil, fmt.Errorf("unsupported backend type %s", storageType)
							}

							kvStore := inputs.TransientStoreService.OpenTransientStore(ctx)
							return ormtable.NewBackend(ormtable.BackendOptions{
								CommitmentStore: kvStore,
								IndexStore:      kvStore,
								ValidateHooks:   nil,
								WriteHooks:      nil,
							}), nil
						}, nil
					default:
						return nil, fmt.Errorf("unsupported backend type %s", storageType)
					}
				},
			})
		}
	}

	return nil, fmt.Errorf("unable to find config for module %s", moduleKey.Name())
}
