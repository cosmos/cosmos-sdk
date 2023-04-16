package flag

import (
	"github.com/spf13/pflag"
	"strconv"
)

func newInt32ToUint64[K int32, V uint64](val map[K]V, p *map[K]V) *genericMapValue[K, V] {
	newInt32Uint64Map := newGenericMapValue(val, p)
	newInt32Uint64Map.Options = genericMapValueOptions[K, V]{
		genericType: "int32ToUint64",
		keyParser: func(s string) (K, error) {
			value, err := strconv.ParseInt(s, 10, 32)
			return K(value), err
		},
		valueParser: func(s string) (V, error) {
			value, err := strconv.ParseUint(s, 10, 64)
			return V(value), err
		},
	}
	return newInt32Uint64Map
}

func Int32ToUint64P(flagSet *pflag.FlagSet, name, shorthand string, value map[int32]uint64, usage string) *map[int32]uint64 {
	p := make(map[int32]uint64)
	flagSet.VarP(newInt32ToUint64(value, &p), name, shorthand, usage)
	return &p
}
