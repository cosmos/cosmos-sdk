package flag

import (
	"github.com/spf13/pflag"
	"strconv"
)

func newStringToBool[K string, V bool](val map[K]V, p *map[K]V) *genericMapValue[K, V] {
	newStringBoolMap := newGenericMapValue(val, p)
	newStringBoolMap.Options = genericMapValueOptions[K, V]{
		genericType: "stringToBool",
		keyParser: func(s string) (K, error) {
			return K(s), nil
		},
		valueParser: func(s string) (V, error) {
			value, err := strconv.ParseBool(s)
			return V(value), err
		},
	}
	return newStringBoolMap
}

func StringToBoolP(flagSet *pflag.FlagSet, name, shorthand string, value map[string]bool, usage string) *map[string]bool {
	p := make(map[string]bool)
	flagSet.VarP(newStringToBool(value, &p), name, shorthand, usage)
	return &p
}
