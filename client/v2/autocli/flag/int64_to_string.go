package flag

import (
	"github.com/spf13/pflag"
	"strconv"
)

func newInt64ToString[K int64, V string](val map[K]V, p *map[K]V) *genericMapValue[K, V] {
	newInt64StringMap := newGenericMapValue(val, p)
	newInt64StringMap.Options = genericMapValueOptions[K, V]{
		genericType: "int64ToString",
		keyParser: func(s string) (K, error) {
			value, err := strconv.ParseInt(s, 10, 64)
			return K(value), err
		},
		valueParser: func(s string) (V, error) {
			return V(s), nil
		},
	}
	return newInt64StringMap
}

func Int64ToStringP(flagSet *pflag.FlagSet, name, shorthand string, value map[int64]string, usage string) *map[int64]string {
	p := make(map[int64]string)
	flagSet.VarP(newInt64ToString(value, &p), name, shorthand, usage)
	return &p
}
