package flag

import (
	"github.com/spf13/pflag"
	"strconv"
)

func newInt64ToInt64[K int64, V int64](val map[K]V, p *map[K]V) *genericMapValue[K, V] {
	newInt64Int64Map := newGenericMapValue(val, p)
	newInt64Int64Map.Options = genericMapValueOptions[K, V]{
		genericType: "int64ToInt64",
		keyParser: func(s string) (K, error) {
			value, err := strconv.ParseInt(s, 10, 64)
			return K(value), err
		},
		valueParser: func(s string) (V, error) {
			value, err := strconv.ParseInt(s, 10, 64)
			return V(value), err
		},
	}
	return newInt64Int64Map
}

func Int64ToInt64P(flagSet *pflag.FlagSet, name, shorthand string, value map[int64]int64, usage string) *map[int64]int64 {
	p := make(map[int64]int64)
	flagSet.VarP(newInt64ToInt64(value, &p), name, shorthand, usage)
	return &p
}
