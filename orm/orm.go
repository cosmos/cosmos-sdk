package orm

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"

	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	modulev1alpha1 "cosmossdk.io/api/cosmos/orm/module/v1alpha1"
	ormv1alpha1 "cosmossdk.io/api/cosmos/orm/v1alpha1"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/orm/model/ormdb"
	"cosmossdk.io/orm/model/ormtable"
)

func init() {
	appconfig.RegisterModule(&modulev1alpha1.Module{},
		appconfig.Provide(ProvideModuleDB),
	)
}

// ModuleDBInputs are the inputs to ProvideModuleDB. NOTE: this is intended to be used by depinject.
type ModuleDBInputs struct {
	depinject.In

	AppConfig             *appv1alpha1.Config
	KVStoreService        store.KVStoreService
	MemoryStoreService    store.MemoryStoreService    `optional:"true"`
	TransientStoreService store.TransientStoreService `optional:"true"`
	TypeResolver          ormtable.TypeResolver       `optional:"true"`
	FileResolver          protodesc.Resolver          `optional:"true"`
}

// ProvideModuleDB provides an ORM ModuleDB scoped to a module. NOTE: this is intended to be used by depinject.
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
				TypeResolver:          inputs.TypeResolver,
				FileResolver:          inputs.FileResolver,
				KVStoreService:        inputs.KVStoreService,
				MemoryStoreService:    inputs.MemoryStoreService,
				TransientStoreService: inputs.TransientStoreService,
			})
		}
	}

	return nil, fmt.Errorf("unable to find config for module %s", moduleKey.Name())
}
