package container

import (
	"reflect"
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

func expandStructArgsConstructor(constructor ConstructorInfo) (ConstructorInfo, error) {
	var foundStructArgs bool
	var newIn []Input

	for _, in := range constructor.In {
		if in.Type.AssignableTo(isStructArgsType) {
			foundStructArgs = true
			inTypes := structArgsInTypes(in.Type)
			newIn = append(newIn, inTypes...)
		} else {
			newIn = append(newIn, in)
		}
	}

	var newOut []Output
	for _, out := range constructor.Out {
		if out.Type.AssignableTo(isStructArgsType) {
			foundStructArgs = true
			newOut = append(newOut, structArgsOutTypes(out.Type)...)
		} else {
			newOut = append(newOut, out)
		}
	}

	if foundStructArgs {
		return ConstructorInfo{
			In:       newIn,
			Out:      newOut,
			Fn:       expandStructArgsFn(constructor),
			Location: constructor.Location,
		}, nil
	}

	return constructor, nil
}

func expandStructArgsFn(constructor ConstructorInfo) func(inputs []reflect.Value) ([]reflect.Value, error) {
	fn := constructor.Fn
	inParams := constructor.In
	outParams := constructor.Out
	return func(inputs []reflect.Value) ([]reflect.Value, error) {
		j := 0
		inputs1 := make([]reflect.Value, len(inParams))
		for i, in := range inParams {
			if in.Type.AssignableTo(isStructArgsType) {
				v, n := buildStructArgs(in.Type, inputs[j:])
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
			if out.Type.AssignableTo(isStructArgsType) {
				outputs1 = append(outputs1, extractFromStructArgs(out.Type, outputs[i])...)
			} else {
				outputs1 = append(outputs1, outputs[i])
			}
		}

		return outputs1, nil
	}
}

func structArgsInTypes(typ reflect.Type) []Input {
	n := typ.NumField()
	var res []Input
	for i := 0; i < n; i++ {
		f := typ.Field(i)
		if f.Type.AssignableTo(isStructArgsType) {
			continue
		}

		optional := f.Tag.Get("optional") == "true"

		res = append(res, Input{
			Type:     f.Type,
			Optional: optional,
		})
	}
	return res
}

func structArgsOutTypes(typ reflect.Type) []Output {
	n := typ.NumField()
	var res []Output
	for i := 0; i < n; i++ {
		f := typ.Field(i)
		if f.Type.AssignableTo(isStructArgsType) {
			continue
		}

		res = append(res, Output{
			Type: f.Type,
		})
	}
	return res
}

func buildStructArgs(typ reflect.Type, values []reflect.Value) (reflect.Value, int) {
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

func extractFromStructArgs(typ reflect.Type, value reflect.Value) []reflect.Value {
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
