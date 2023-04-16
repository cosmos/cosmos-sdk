package flag

import (
	"github.com/spf13/pflag"
	"strconv"
)

func newInt32ToInt64[K int32, V int64](val map[K]V, p *map[K]V) *genericMapValue[K, V] {
	newInt32Int64Map := newGenericMapValue(val, p)
	newInt32Int64Map.Options = genericMapValueOptions[K, V]{
		genericType: "int32ToInt64",
		keyParser: func(s string) (K, error) {
			value, err := strconv.ParseInt(s, 10, 32)
			return K(value), err
		},
		valueParser: func(s string) (V, error) {
			value, err := strconv.ParseInt(s, 10, 64)
			return V(value), err
		},
	}
	return newInt32Int64Map
}

func Int32ToInt64P(flagSet *pflag.FlagSet, name, shorthand string, value map[int32]int64, usage string) *map[int32]int64 {
	p := make(map[int32]int64)
	flagSet.VarP(newInt32ToInt64(value, &p), name, shorthand, usage)
	return &p
}
