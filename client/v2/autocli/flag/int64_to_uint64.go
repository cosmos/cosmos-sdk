package flag

import (
	"github.com/spf13/pflag"
	"strconv"
)

func newInt64ToUint64[K int64, V uint64](val map[K]V, p *map[K]V) *genericMapValue[K, V] {
	newInt64Uint64Map := newGenericMapValue(val, p)
	newInt64Uint64Map.Options = genericMapValueOptions[K, V]{
		genericType: "int64ToUint64",
		keyParser: func(s string) (K, error) {
			value, err := strconv.ParseInt(s, 10, 64)
			return K(value), err
		},
		valueParser: func(s string) (V, error) {
			value, err := strconv.ParseUint(s, 10, 64)
			return V(value), err
		},
	}
	return newInt64Uint64Map
}

func Int64ToUint64P(flagSet *pflag.FlagSet, name, shorthand string, value map[int64]uint64, usage string) *map[int64]uint64 {
	p := make(map[int64]uint64)
	flagSet.VarP(newInt64ToUint64(value, &p), name, shorthand, usage)
	return &p
}
