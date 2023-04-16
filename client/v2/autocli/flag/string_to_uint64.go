package flag

import (
	"github.com/spf13/pflag"
	"strconv"
)

func newStringToUint64[K string, V uint64](val map[K]V, p *map[K]V) *genericMapValue[K, V] {
	newStringUint64Map := newGenericMapValue(val, p)
	newStringUint64Map.Options = genericMapValueOptions[K, V]{
		genericType: "stringToUint64",
		keyParser: func(s string) (K, error) {
			return K(s), nil
		},
		valueParser: func(s string) (V, error) {
			value, err := strconv.ParseUint(s, 10, 64)
			return V(value), err
		},
	}
	return newStringUint64Map
}

func StringToUint64P(flagSet *pflag.FlagSet, name, shorthand string, value map[string]uint64, usage string) *map[string]uint64 {
	p := make(map[string]uint64)
	flagSet.VarP(newStringToUint64(value, &p), name, shorthand, usage)
	return &p
}
