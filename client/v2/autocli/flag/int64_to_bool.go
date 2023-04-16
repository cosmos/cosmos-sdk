package flag

import (
	"github.com/spf13/pflag"
	"strconv"
)

func newInt64ToBool[K int64, V bool](val map[K]V, p *map[K]V) *genericMapValue[K, V] {
	newInt64BoolMap := newGenericMapValue(val, p)
	newInt64BoolMap.Options = genericMapValueOptions[K, V]{
		genericType: "int64ToBool",
		keyParser: func(s string) (K, error) {
			value, err := strconv.ParseInt(s, 10, 64)
			return K(value), err
		},
		valueParser: func(s string) (V, error) {
			value, err := strconv.ParseBool(s)
			return V(value), err
		},
	}
	return newInt64BoolMap
}

func Int64ToBoolP(flagSet *pflag.FlagSet, name, shorthand string, value map[int64]bool, usage string) *map[int64]bool {
	p := make(map[int64]bool)
	flagSet.VarP(newInt64ToBool(value, &p), name, shorthand, usage)
	return &p
}
