package depinject

import (
	"reflect"

	"github.com/pkg/errors"
)

// providerDescriptor defines a special provider type that is defined by
// reflection. It should be passed as a value to the Provide function.
// Ex:
//   option.Provide(providerDescriptor{ ... })
type providerDescriptor struct {
	// Inputs defines the in parameter types to Fn.
	Inputs []providerInput

	// Outputs defines the out parameter types to Fn.
	Outputs []providerOutput

	// Fn defines the provider function.
	Fn func([]reflect.Value) ([]reflect.Value, error)

	// Location defines the source code location to be used for this provider
	// in error messages.
	Location Location
}

type providerInput struct {
	Type     reflect.Type
	Optional bool
}

type providerOutput struct {
	Type reflect.Type
}

func extractProviderDescriptor(provider interface{}) (providerDescriptor, error) {
	rctr, err := doExtractProviderDescriptor(provider)
	if err != nil {
		return providerDescriptor{}, err
	}
	return expandStructArgsProvider(rctr)
}

func extractInvokerDescriptor(provider interface{}) (providerDescriptor, error) {
	var err error
	rctr, err := doExtractProviderDescriptor(provider)

	// mark all inputs as optional
	for i, input := range rctr.Inputs {
		input.Optional = true
		rctr.Inputs[i] = input
	}

	if err != nil {
		return providerDescriptor{}, err
	}
	return expandStructArgsProvider(rctr)
}

func doExtractProviderDescriptor(ctr interface{}) (providerDescriptor, error) {
	val := reflect.ValueOf(ctr)
	typ := val.Type()
	if typ.Kind() != reflect.Func {
		return providerDescriptor{}, errors.Errorf("expected a Func type, got %v", typ)
	}

	loc := LocationFromPC(val.Pointer())

	if typ.IsVariadic() {
		return providerDescriptor{}, errors.Errorf("variadic function can't be used as a provider: %s", loc)
	}

	numIn := typ.NumIn()
	in := make([]providerInput, numIn)
	for i := 0; i < numIn; i++ {
		in[i] = providerInput{
			Type: typ.In(i),
		}
	}

	errIdx := -1
	numOut := typ.NumOut()
	var out []providerOutput
	for i := 0; i < numOut; i++ {
		t := typ.Out(i)
		if t == errType {
			if i != numOut-1 {
				return providerDescriptor{}, errors.Errorf("output error parameter is not last parameter in function %s", loc)
			}
			errIdx = i
		} else {
			out = append(out, providerOutput{Type: t})
		}
	}

	return providerDescriptor{
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
