package container

import (
	"reflect"

	containerreflect "github.com/cosmos/cosmos-sdk/container/reflect"
)

// StructArgs is a type which can be embedded in another struct to alert the
// container that the fields of the struct are dependency inputs/outputs. That
// is, the container will not look to resolve a value with StructArgs embedded
// directly, but will instead use the struct's fields to resolve or populate
// dependencies. Types with embedded StructArgs can be used in both the input
// and output parameter positions.
type StructArgs struct{}

func (StructArgs) isStructArgs() {}

type isStructArgs interface {
	isStructArgs()
}

var isStructArgsType = reflect.TypeOf((*isStructArgs)(nil)).Elem()

func expandStructArgsConstructor(constructor *containerreflect.Constructor) (*containerreflect.Constructor, error) {
	var foundStructArgs bool
	var newIn []containerreflect.Input

	for _, in := range constructor.In {
		if in.Type.AssignableTo(isStructArgsType) {
			foundStructArgs = true
			inTypes := structArgsInTypes(in.Type)
			newIn = append(newIn, inTypes...)
		} else {
			newIn = append(newIn, in)
		}
	}

	var newOut []containerreflect.Output
	for _, out := range constructor.Out {
		if out.Type.AssignableTo(isStructArgsType) {
			foundStructArgs = true
			newOut = append(newOut, structArgsOutTypes(out.Type)...)
		} else {
			newOut = append(newOut, out)
		}
	}

	if foundStructArgs {
		return &containerreflect.Constructor{
			In:       newIn,
			Out:      newOut,
			Fn:       expandStructArgsFn(constructor),
			Location: constructor.Location,
		}, nil
	}

	return constructor, nil
}

func expandStructArgsFn(constructor *containerreflect.Constructor) func(inputs []reflect.Value) ([]reflect.Value, error) {
	fn := constructor.Fn
	inParams := constructor.In
	outParams := constructor.Out
	return func(inputs []reflect.Value) ([]reflect.Value, error) {
		j := 0
		inputs1 := make([]reflect.Value, len(constructor.In))
		for i, in := range inParams {
			if in.Type.AssignableTo(isStructArgsType) {
				v, n := makeStructArgs(in.Type, inputs[j:])
				inputs1[i] = v
				j += n
			} else {
				inputs1 = append(inputs1, inputs[j])
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
			if out.Type.AssignableTo(isStructArgsType) {
				outputs1 = append(outputs1, extractStructArgs(out.Type, outputs[i])...)
			} else {
				outputs1 = append(outputs1, outputs[i])
			}
		}

		return outputs1, nil
	}
}

func structArgsInTypes(typ reflect.Type) []containerreflect.Input {
	n := typ.NumField()
	var res []containerreflect.Input
	for i := 0; i < n; i++ {
		f := typ.Field(i)
		if f.Type.AssignableTo(isStructArgsType) {
			continue
		}

		optional := f.Tag.Get("optional") == "true"

		res = append(res, containerreflect.Input{
			Type:     f.Type,
			Optional: optional,
		})
	}
	return res
}

func structArgsOutTypes(typ reflect.Type) []containerreflect.Output {
	n := typ.NumField()
	var res []containerreflect.Output
	for i := 0; i < n; i++ {
		f := typ.Field(i)
		if f.Type.AssignableTo(isStructArgsType) {
			continue
		}

		res = append(res, containerreflect.Output{
			Type: f.Type,
		})
	}
	return res
}

func makeStructArgs(typ reflect.Type, values []reflect.Value) (reflect.Value, int) {
	numFields := typ.NumField()
	j := 0
	res := reflect.New(typ)
	for i := 0; i < numFields; i++ {
		f := typ.Field(i)
		if f.Type.AssignableTo(isStructArgsType) {
			continue
		}

		res.Elem().Field(i).Set(values[j])
		j++
	}
	return res.Elem(), j
}

func extractStructArgs(typ reflect.Type, value reflect.Value) []reflect.Value {
	numFields := typ.NumField()
	var res []reflect.Value
	for i := 0; i < numFields; i++ {
		f := typ.Field(i)
		if f.Type.AssignableTo(isStructArgsType) {
			continue
		}

		res = append(res, value.Field(i))
	}
	return res
}
