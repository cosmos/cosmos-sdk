package flag

import (
	"github.com/spf13/pflag"
	"strconv"
)

func newStringToUint32[K string, V uint32](val map[K]V, p *map[K]V) *genericMapValue[K, V] {
	newStringUintMap := newGenericMapValue(val, p)
	newStringUintMap.Options = genericMapValueOptions[K, V]{
		genericType: "stringToUint32",
		keyParser: func(s string) (K, error) {
			return K(s), nil
		},
		valueParser: func(s string) (V, error) {
			value, err := strconv.ParseUint(s, 10, 32)
			return V(value), err
		},
	}
	return newStringUintMap
}

func StringToUint32P(flagSet *pflag.FlagSet, name, shorthand string, value map[string]uint32, usage string) *map[string]uint32 {
	p := make(map[string]uint32)
	flagSet.VarP(newStringToUint32(value, &p), name, shorthand, usage)
	return &p
}
