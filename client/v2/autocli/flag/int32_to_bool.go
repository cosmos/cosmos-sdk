package flag

import (
	"github.com/spf13/pflag"
	"strconv"
)

func newInt32ToBool[K int32, V bool](val map[K]V, p *map[K]V) *genericMapValue[K, V] {
	newInt32BoolMap := newGenericMapValue(val, p)
	newInt32BoolMap.Options = genericMapValueOptions[K, V]{
		genericType: "int32ToBool",
		keyParser: func(s string) (K, error) {
			value, err := strconv.ParseInt(s, 10, 32)
			return K(value), err
		},
		valueParser: func(s string) (V, error) {
			value, err := strconv.ParseBool(s)
			return V(value), err
		},
	}
	return newInt32BoolMap
}

func Int32ToBoolP(flagSet *pflag.FlagSet, name, shorthand string, value map[int32]bool, usage string) *map[int32]bool {
	p := make(map[int32]bool)
	flagSet.VarP(newInt32ToBool(value, &p), name, shorthand, usage)
	return &p
}
