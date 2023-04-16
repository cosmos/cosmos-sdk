package flag

import (
	"github.com/spf13/pflag"
	"strconv"
)

func newInt32ToUint32[K int32, V uint32](val map[K]V, p *map[K]V) *genericMapValue[K, V] {
	newInt32Uint32Map := newGenericMapValue(val, p)
	newInt32Uint32Map.Options = genericMapValueOptions[K, V]{
		genericType: "int32ToUint32",
		keyParser: func(s string) (K, error) {
			value, err := strconv.ParseInt(s, 10, 32)
			return K(value), err
		},
		valueParser: func(s string) (V, error) {
			value, err := strconv.ParseUint(s, 10, 32)
			return V(value), err
		},
	}
	return newInt32Uint32Map
}

func Int32ToUint32P(flagSet *pflag.FlagSet, name, shorthand string, value map[int32]uint32, usage string) *map[int32]uint32 {
	p := make(map[int32]uint32)
	flagSet.VarP(newInt32ToUint32(value, &p), name, shorthand, usage)
	return &p
}
