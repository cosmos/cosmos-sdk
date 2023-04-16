package flag

import (
	"github.com/spf13/pflag"
	"strconv"
)

func newInt64ToUint32[K int64, V uint32](val map[K]V, p *map[K]V) *genericMapValue[K, V] {
	newInt64Uint32Map := newGenericMapValue(val, p)
	newInt64Uint32Map.Options = genericMapValueOptions[K, V]{
		genericType: "int64ToUint32",
		keyParser: func(s string) (K, error) {
			value, err := strconv.ParseInt(s, 10, 64)
			return K(value), err
		},
		valueParser: func(s string) (V, error) {
			value, err := strconv.ParseUint(s, 10, 32)
			return V(value), err
		},
	}
	return newInt64Uint32Map
}

func Int64ToUint32P(flagSet *pflag.FlagSet, name, shorthand string, value map[int64]uint32, usage string) *map[int64]uint32 {
	p := make(map[int64]uint32)
	flagSet.VarP(newInt64ToUint32(value, &p), name, shorthand, usage)
	return &p
}
