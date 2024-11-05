package benchmark

import (
	"unsafe"

	modulev1 "cosmossdk.io/api/cosmos/benchmark/module/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject/appconfig"
)

const ModuleName = "benchmark"

func init() {
	// TODO try depinject gogo API
	appconfig.RegisterModule(
		&modulev1.Module{},
		appconfig.Provide(
			ProvideModule,
		),
	)
}

func ProvideModule(collector *KVServiceCollector) appmodule.AppModule {
	return AppModule{
		collector: collector,
	}
}

type KVServiceCollector struct {
	services map[string]store.KVStoreService
}

func NewKVServiceCollector() *KVServiceCollector {
	return &KVServiceCollector{
		services: make(map[string]store.KVStoreService),
	}
}

func NewKVStoreServiceFactory(
	collector *KVServiceCollector,
	runtimeFactory store.KVStoreServiceFactory,
) store.KVStoreServiceFactory {
	return func(address []byte) store.KVStoreService {
		service := runtimeFactory(address)
		collector.services[unsafeBytesToStr(address)] = service
		return service
	}
}

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
