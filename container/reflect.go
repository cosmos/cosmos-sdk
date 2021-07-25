package container

import (
	"reflect"

	"github.com/pkg/errors"

	containerreflect "github.com/cosmos/cosmos-sdk/container/reflect"
)

func makeReflectConstructor(ctr interface{}) (*containerreflect.Constructor, error) {
	rctr, ok := ctr.(containerreflect.Constructor)
	if !ok {
		val := reflect.ValueOf(ctr)
		if val.Type().Kind() != reflect.Func {
			return nil, errors.Errorf("expected constructor function, got %T", ctr)
		}
		loc := containerreflect.LocationFromPC(val.Pointer())
		typ := val.Type()
		if typ.Kind() != reflect.Func {
			return nil, errors.Errorf("expected a Func type, got %T", ctr)
		}

		numIn := typ.NumIn()
		in := make([]containerreflect.Input, numIn)
		for i := 0; i < numIn; i++ {
			in[i] = containerreflect.Input{
				Type: typ.In(i),
			}
		}

		errIdx := -1
		numOut := typ.NumOut()
		var out []containerreflect.Output
		for i := 0; i < numOut; i++ {
			t := typ.Out(i)
			if t == errType {
				if i != numOut-1 {
					return nil, errors.Errorf("output error parameter is not last parameter in function %s", loc)
				}
				errIdx = i
			} else {
				out = append(out, containerreflect.Output{Type: t})
			}
		}

		rctr = containerreflect.Constructor{
			In:  in,
			Out: out,
			Fn: func(values []reflect.Value) ([]reflect.Value, error) {
				res := val.Call(values)
				if errIdx > 0 {
					err := res[errIdx]
					if !err.IsZero() {
						return nil, err.Interface().(error)
					}
					return res[0:errIdx], nil
				}
				return res, nil
			},
			Location: loc,
		}
	}

	return expandStructArgsConstructor(&rctr)
}

var errType = reflect.TypeOf((*error)(nil)).Elem()
