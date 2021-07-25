package container

import (
	"reflect"

	"github.com/pkg/errors"

	containerreflect "github.com/cosmos/cosmos-sdk/container/reflect"
)

func reflectConstructor(ctr interface{}) (containerreflect.Constructor, error) {
	rctr, ok := ctr.(containerreflect.Constructor)
	if !ok {
		val := reflect.ValueOf(ctr)
		mctr, haveMethodCtr := ctr.(containerreflect.MethodConstructor)
		var instance interface{}
		if haveMethodCtr {
			val = mctr.Method.Func
			instance = mctr.Instance
		}

		var err error
		rctr, err = makeReflectConstructor(val)
		if err != nil {
			return containerreflect.Constructor{}, err
		}

		if haveMethodCtr {
			rctr = adaptMethodConstructor(rctr, instance)
		}
	}

	return expandStructArgsConstructor(rctr)
}

func makeReflectConstructor(val reflect.Value) (containerreflect.Constructor, error) {
	loc := containerreflect.LocationFromPC(val.Pointer())
	typ := val.Type()
	if typ.Kind() != reflect.Func {
		return containerreflect.Constructor{}, errors.Errorf("expected a Func type, got %v", typ)
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
				return containerreflect.Constructor{}, errors.Errorf("output error parameter is not last parameter in function %s", loc)
			}
			errIdx = i
		} else {
			out = append(out, containerreflect.Output{Type: t})
		}
	}

	return containerreflect.Constructor{
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
	}, nil
}

func adaptMethodConstructor(rctr containerreflect.Constructor, instance interface{}) containerreflect.Constructor {
	rctr.In = rctr.In[1:]
	fn := rctr.Fn
	instanceValSlice := []reflect.Value{reflect.ValueOf(instance)}
	rctr.Fn = func(values []reflect.Value) ([]reflect.Value, error) {
		values = append(instanceValSlice, values...)
		return fn(values)
	}
	return rctr
}

var errType = reflect.TypeOf((*error)(nil)).Elem()
