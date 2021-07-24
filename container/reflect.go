package container

import (
	"fmt"
	"reflect"

	containerreflect "github.com/cosmos/cosmos-sdk/container/reflect"
)

func makeReflectConstructor(ctr interface{}) (*containerreflect.Constructor, error) {
	rctr, ok := ctr.(containerreflect.Constructor)
	if !ok {
		val := reflect.ValueOf(ctr)
		typ := val.Type()
		if typ.Kind() != reflect.Func {
			return nil, fmt.Errorf("expected a Func type, got %T", ctr)
		}

		numIn := typ.NumIn()
		in := make([]containerreflect.Input, numIn)
		for i := 0; i < numIn; i++ {
			in[i] = containerreflect.Input{
				Type: typ.In(i),
			}
		}

		numOut := typ.NumOut()
		out := make([]containerreflect.Output, numOut)
		for i := 0; i < numOut; i++ {
			out[i] = containerreflect.Output{Type: typ.Out(i)}
		}

		rctr = containerreflect.Constructor{
			In:  in,
			Out: out,
			Fn: func(values []reflect.Value) []reflect.Value {
				return val.Call(values)
			},
			Location: containerreflect.LocationFromPC(val.Pointer()),
		}
	}

	return &rctr, nil
}
