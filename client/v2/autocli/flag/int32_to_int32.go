package flag

import (
	"github.com/spf13/pflag"
	"strconv"
)

func newInt32ToInt32[K int32, V int32](val map[K]V, p *map[K]V) *genericMapValue[K, V] {
	newInt32Int32Map := newGenericMapValue(val, p)
	newInt32Int32Map.Options = genericMapValueOptions[K, V]{
		genericType: "int32ToInt32",
		keyParser: func(s string) (K, error) {
			value, err := strconv.ParseInt(s, 10, 32)
			return K(value), err
		},
		valueParser: func(s string) (V, error) {
			value, err := strconv.ParseInt(s, 10, 32)
			return V(value), err
		},
	}
	return newInt32Int32Map
}

func Int32ToInt32P(flagSet *pflag.FlagSet, name, shorthand string, value map[int32]int32, usage string) *map[int32]int32 {
	p := make(map[int32]int32)
	flagSet.VarP(newInt32ToInt32(value, &p), name, shorthand, usage)
	return &p
}
