package module

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"unicode"
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

type StoreKeyRegistrar interface {
	RegisterKey(string)
}

func ProvideModule(
	cfg *modulev1.Module,
	registrar StoreKeyRegistrar,
	kvStoreServiceFactory store.KVStoreServiceFactory,
) (appmodule.AppModule, error) {
	rand := rand.New(rand.NewSource(int64(cfg.GenesisParams.Seed)))
	collector := NewKVServiceCollector()
	for range cfg.GenesisParams.StoreKeyCount {
		suffix, err := randomWords(rand, 2, "_")
		if err != nil {
			return nil, err
		}
		sk := fmt.Sprintf("bench_%s", suffix)
		registrar.RegisterKey(sk)
		kvService := kvStoreServiceFactory(unsafeStrToBytes(sk))
		collector.services[sk] = kvService
	}
	return NewAppModule(collector), nil
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

func randomWords(rand *rand.Rand, count int, separator string) (string, error) {
	const wordsPath = "/usr/share/dict/words"
	stat, err := os.Stat(wordsPath)
	if err != nil {
		return "", err
	}
	f, err := os.Open(wordsPath)
	if err != nil {
		return "", err
	}

	words := make([]string, count)
	buf := make([]byte, 1)
	for i := 0; i < count; i++ {
		x := rand.Intn(int(stat.Size()))
		_, err := f.Seek(int64(x), 0)
		if err != nil {
			return "", err
		}
		for {
			_, err := f.Read(buf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					i--
					continue
				}
			}
			if string(buf) == "\n" {
				break
			}
		}
		for {
			_, err := f.Read(buf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
			}
			if string(buf) == "\n" {
				break
			}
			words[i] += string(buf)
		}
		if len(words[i]) < 4 || len(words[i]) > 8 || unicode.IsUpper(rune(words[i][0])) {
			words[i] = ""
			i--
			continue
		}
	}
	return strings.Join(words, separator), nil
}
