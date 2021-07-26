package container

import (
	"reflect"

	"github.com/pkg/errors"
)

// ConstructorInfo defines a special constructor type that is defined by
// reflection. It should be passed as a value to the Provide function.
// Ex:
//   option.Provide(ConstructorInfo{ ... })
type ConstructorInfo struct {
	// In defines the in parameter types to Fn.
	In []Input

	// Out defines the out parameter types to Fn.
	Out []Output

	// Fn defines the constructor function.
	Fn func([]reflect.Value) ([]reflect.Value, error)

	// Location defines the source code location to be used for this constructor
	// in error messages.
	Location Location
}

type Input struct {
	Type     reflect.Type
	Optional bool
}

type Output struct {
	Type reflect.Type
}

func getConstructorInfo(ctr interface{}) (ConstructorInfo, error) {
	rctr, ok := ctr.(ConstructorInfo)
	if !ok {
		var err error
		rctr, err = ConstructorInfoFor(ctr)
		if err != nil {
			return ConstructorInfo{}, err
		}
	}

	return expandStructArgsConstructor(rctr)
}

func ConstructorInfoFor(ctr interface{}) (ConstructorInfo, error) {
	val := reflect.ValueOf(ctr)
	typ := val.Type()
	if typ.Kind() != reflect.Func {
		return ConstructorInfo{}, errors.Errorf("expected a Func type, got %v", typ)
	}

	loc := LocationFromPC(val.Pointer())
	numIn := typ.NumIn()
	in := make([]Input, numIn)
	for i := 0; i < numIn; i++ {
		in[i] = Input{
			Type: typ.In(i),
		}
	}

	errIdx := -1
	numOut := typ.NumOut()
	var out []Output
	for i := 0; i < numOut; i++ {
		t := typ.Out(i)
		if t == errType {
			if i != numOut-1 {
				return ConstructorInfo{}, errors.Errorf("output error parameter is not last parameter in function %s", loc)
			}
			errIdx = i
		} else {
			out = append(out, Output{Type: t})
		}
	}

	return ConstructorInfo{
		In:  in,
		Out: out,
		Fn: func(values []reflect.Value) ([]reflect.Value, error) {
			res := val.Call(values)
			if errIdx >= 0 {
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

var errType = reflect.TypeOf((*error)(nil)).Elem()
