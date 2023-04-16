package flag

import (
	"github.com/spf13/pflag"
	"strconv"
)

func newInt64ToInt32[K int64, V int32](val map[K]V, p *map[K]V) *genericMapValue[K, V] {
	newInt64Int32Map := newGenericMapValue(val, p)
	newInt64Int32Map.Options = genericMapValueOptions[K, V]{
		genericType: "int64ToInt32",
		keyParser: func(s string) (K, error) {
			value, err := strconv.ParseInt(s, 10, 64)
			return K(value), err
		},
		valueParser: func(s string) (V, error) {
			value, err := strconv.ParseInt(s, 10, 32)
			return V(value), err
		},
	}
	return newInt64Int32Map
}

func Int64ToInt32P(flagSet *pflag.FlagSet, name, shorthand string, value map[int64]int32, usage string) *map[int64]int32 {
	p := make(map[int64]int32)
	flagSet.VarP(newInt64ToInt32(value, &p), name, shorthand, usage)
	return &p
}
