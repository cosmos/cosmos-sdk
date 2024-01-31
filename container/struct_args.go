package container

import (
	"reflect"

	"github.com/pkg/errors"
)

// In can be embedded in another struct to inform the container that the
// fields of the struct should be treated as dependency inputs.
// This allows a struct to be used to specify dependencies rather than
// positional parameters.
//
// Fields of the struct may support the following tags:
//
//	optional	if set to true, the dependency is optional and will
//				be set to its default value if not found, rather than causing
//				an error
type In struct{}

func (In) isIn() {}

type isIn interface{ isIn() }

var isInType = reflect.TypeOf((*isIn)(nil)).Elem()

// Out can be embedded in another struct to inform the container that the
// fields of the struct should be treated as dependency outputs.
// This allows a struct to be used to specify outputs rather than
// positional return values.
type Out struct{}

func (Out) isOut() {}

type isOut interface{ isOut() }

var isOutType = reflect.TypeOf((*isOut)(nil)).Elem()

func expandStructArgsProvider(provider ProviderDescriptor) (ProviderDescriptor, error) {
	var structArgsInInput bool
	var newIn []ProviderInput
	for _, in := range provider.Inputs {
		if in.Type.AssignableTo(isInType) {
			structArgsInInput = true
			inTypes, err := structArgsInTypes(in.Type)
			if err != nil {
				return ProviderDescriptor{}, err
			}
			newIn = append(newIn, inTypes...)
		} else {
			newIn = append(newIn, in)
		}
	}

	newOut, structArgsInOutput := expandStructArgsOutTypes(provider.Outputs)

	if structArgsInInput || structArgsInOutput {
		return ProviderDescriptor{
			Inputs:   newIn,
			Outputs:  newOut,
			Fn:       expandStructArgsFn(provider),
			Location: provider.Location,
		}, nil
	}

	return provider, nil
}

func expandStructArgsFn(provider ProviderDescriptor) func(inputs []reflect.Value) ([]reflect.Value, error) {
	fn := provider.Fn
	inParams := provider.Inputs
	outParams := provider.Outputs
	return func(inputs []reflect.Value) ([]reflect.Value, error) {
		j := 0
		inputs1 := make([]reflect.Value, len(inParams))
		for i, in := range inParams {
			if in.Type.AssignableTo(isInType) {
				v, n := buildIn(in.Type, inputs[j:])
				inputs1[i] = v
				j += n
			} else {
				inputs1[i] = inputs[j]
				j++
			}
		}

		outputs, err := fn(inputs1)
		if err != nil {
			return nil, err
		}

		var outputs1 []reflect.Value
		for i, out := range outParams {
			if out.Type.AssignableTo(isOutType) {
				outputs1 = append(outputs1, extractFromOut(out.Type, outputs[i])...)
			} else {
				outputs1 = append(outputs1, outputs[i])
			}
		}

		return outputs1, nil
	}
}

func structArgsInTypes(typ reflect.Type) ([]ProviderInput, error) {
	n := typ.NumField()
	var res []ProviderInput
	for i := 0; i < n; i++ {
		f := typ.Field(i)
		if f.Type.AssignableTo(isInType) {
			continue
		}

		var optional bool
		optTag, found := f.Tag.Lookup("optional")
		if found {
			if optTag == "true" {
				optional = true
			} else {
				return nil, errors.Errorf("bad optional tag %q (should be \"true\") in %v", optTag, typ)
			}
		}

		res = append(res, ProviderInput{
			Type:     f.Type,
			Optional: optional,
		})
	}
	return res, nil
}

func expandStructArgsOutTypes(outputs []ProviderOutput) ([]ProviderOutput, bool) {
	foundStructArgs := false
	var newOut []ProviderOutput
	for _, out := range outputs {
		if out.Type.AssignableTo(isOutType) {
			foundStructArgs = true
			newOut = append(newOut, structArgsOutTypes(out.Type)...)
		} else {
			newOut = append(newOut, out)
		}
	}
	return newOut, foundStructArgs
}

func structArgsOutTypes(typ reflect.Type) []ProviderOutput {
	n := typ.NumField()
	var res []ProviderOutput
	for i := 0; i < n; i++ {
		f := typ.Field(i)
		if f.Type.AssignableTo(isOutType) {
			continue
		}

		res = append(res, ProviderOutput{
			Type: f.Type,
		})
	}
	return res
}

func buildIn(typ reflect.Type, values []reflect.Value) (reflect.Value, int) {
	numFields := typ.NumField()
	j := 0
	res := reflect.New(typ)
	for i := 0; i < numFields; i++ {
		f := typ.Field(i)
		if f.Type.AssignableTo(isInType) {
			continue
		}

		res.Elem().Field(i).Set(values[j])
		j++
	}
	return res.Elem(), j
}

func extractFromOut(typ reflect.Type, value reflect.Value) []reflect.Value {
	numFields := typ.NumField()
	var res []reflect.Value
	for i := 0; i < numFields; i++ {
		f := typ.Field(i)
		if f.Type.AssignableTo(isOutType) {
			continue
		}

		res = append(res, value.Field(i))
	}
	return res
}
