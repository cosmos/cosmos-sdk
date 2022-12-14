package container

import (
	"reflect"

	"github.com/pkg/errors"
)

// ProviderDescriptor defines a special provider type that is defined by
// reflection. It should be passed as a value to the Provide function.
// Ex:
//
//	option.Provide(ProviderDescriptor{ ... })
type ProviderDescriptor struct {
	// Inputs defines the in parameter types to Fn.
	Inputs []ProviderInput

	// Outputs defines the out parameter types to Fn.
	Outputs []ProviderOutput

	// Fn defines the provider function.
	Fn func([]reflect.Value) ([]reflect.Value, error)

	// Location defines the source code location to be used for this provider
	// in error messages.
	Location Location
}

type ProviderInput struct {
	Type     reflect.Type
	Optional bool
}

type ProviderOutput struct {
	Type reflect.Type
}

func ExtractProviderDescriptor(provider interface{}) (ProviderDescriptor, error) {
	rctr, ok := provider.(ProviderDescriptor)
	if !ok {
		var err error
		rctr, err = doExtractProviderDescriptor(provider)
		if err != nil {
			return ProviderDescriptor{}, err
		}
	}

	return expandStructArgsProvider(rctr)
}

func doExtractProviderDescriptor(ctr interface{}) (ProviderDescriptor, error) {
	val := reflect.ValueOf(ctr)
	typ := val.Type()
	if typ.Kind() != reflect.Func {
		return ProviderDescriptor{}, errors.Errorf("expected a Func type, got %v", typ)
	}

	loc := LocationFromPC(val.Pointer())

	if typ.IsVariadic() {
		return ProviderDescriptor{}, errors.Errorf("variadic function can't be used as a provider: %s", loc)
	}

	numIn := typ.NumIn()
	in := make([]ProviderInput, numIn)
	for i := 0; i < numIn; i++ {
		in[i] = ProviderInput{
			Type: typ.In(i),
		}
	}

	errIdx := -1
	numOut := typ.NumOut()
	var out []ProviderOutput
	for i := 0; i < numOut; i++ {
		t := typ.Out(i)
		if t == errType {
			if i != numOut-1 {
				return ProviderDescriptor{}, errors.Errorf("output error parameter is not last parameter in function %s", loc)
			}
			errIdx = i
		} else {
			out = append(out, ProviderOutput{Type: t})
		}
	}

	return ProviderDescriptor{
		Inputs:  in,
		Outputs: out,
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
