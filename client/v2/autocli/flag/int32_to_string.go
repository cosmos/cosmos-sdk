package flag

import (
	"github.com/spf13/pflag"
	"strconv"
)

func newInt32ToString[K int32, V string](val map[K]V, p *map[K]V) *genericMapValue[K, V] {
	newInt32StringMap := newGenericMapValue(val, p)
	newInt32StringMap.Options = genericMapValueOptions[K, V]{
		genericType: "int32ToString",
		keyParser: func(s string) (K, error) {
			value, err := strconv.ParseInt(s, 10, 32)
			return K(value), err
		},
		valueParser: func(s string) (V, error) {
			return V(s), nil
		},
	}
	return newInt32StringMap
}

func Int32ToStringP(flagSet *pflag.FlagSet, name, shorthand string, value map[int32]string, usage string) *map[int32]string {
	p := make(map[int32]string)
	flagSet.VarP(newInt32ToString(value, &p), name, shorthand, usage)
	return &p
}
