package module

import (
	"fmt"
	"unsafe"

	modulev1 "cosmossdk.io/api/cosmos/benchmark/module/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/log"
	gen "cosmossdk.io/x/benchmark/generator"
)

const ModuleName = "benchmark"
const maxStoreKeyGenIterations = 100

func init() {
	// TODO try depinject gogo API
	appconfig.RegisterModule(
		&modulev1.Module{},
		appconfig.Provide(
			ProvideModule,
		),
	)
}

type StoreKeyRegistrar interface {
	RegisterKey(string)
}

func ProvideModule(
	logger log.Logger,
	cfg *modulev1.Module,
	registrar StoreKeyRegistrar,
	kvStoreServiceFactory store.KVStoreServiceFactory,
) (appmodule.AppModule, error) {
	g := gen.NewGenerator(gen.Options{Seed: cfg.GenesisParams.Seed})
	kvMap := make(KVServiceMap)
	storeKeys := make([]string, cfg.GenesisParams.StoreKeyCount)

	var i, j uint64
	for i < cfg.GenesisParams.StoreKeyCount {
		if j > maxStoreKeyGenIterations {
			return nil, fmt.Errorf("failed to generate %d unique store keys", cfg.GenesisParams.StoreKeyCount)
		}
		sk := fmt.Sprintf("%s_%x", ModuleName, g.Bytes(cfg.GenesisParams.Seed, 8))
		if _, ok := kvMap[sk]; ok {
			j++
			continue
		}
		registrar.RegisterKey(sk)
		kvService := kvStoreServiceFactory(unsafeStrToBytes(sk))
		kvMap[sk] = kvService
		storeKeys[i] = sk
		i++
		j++
	}
	return NewAppModule(cfg.GenesisParams, storeKeys, kvMap, logger), nil
}

type KVServiceMap map[string]store.KVStoreService

// unsafeStrToBytes uses unsafe to convert string into byte array. Returned bytes
// must not be altered after this function is called as it will cause a segmentation fault.
func unsafeStrToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s)) // ref https://github.com/golang/go/issues/53003#issuecomment-1140276077
}

// unsafeBytesToStr is meant to make a zero allocation conversion
// from []byte -> string to speed up operations, it is not meant
// to be used generally, but for a specific pattern to delete keys
// from a map.
func unsafeBytesToStr(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
