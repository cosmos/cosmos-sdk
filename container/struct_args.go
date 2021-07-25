package container

import (
	"fmt"
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

func extractStructArgs(constructor *containerreflect.Constructor) (*containerreflect.Constructor, error) {
	var foundStructArgs bool
	var newIn []containerreflect.Input
	var inSplicers []func([]reflect.Value) ([]reflect.Value, error)

	for i, in := range constructor.In {
		if in.Type.AssignableTo(isStructArgsType) {
			foundStructArgs = true
			inTypes := structArgsInTypes(in.Type)
			newIn = append(newIn, inTypes...)
			inSplicers = append(inSplicers, makeInSplicer(in.Type, i, len(inTypes)))
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
			In:  newIn,
			Out: newOut,
			Fn: func(values []reflect.Value) ([]reflect.Value, error) {
				j := 0
				values1 := make([]reflect.Value, len(constructor.In))
				for i, in := range constructor.In {
					if in.Type.AssignableTo(isStructArgsType) {
						v, n := makeStructArgs(in.Type, values1[j:])
						values[i] = v
						j += n
					} else {
						values1 = append(values1, values[j])
						values1[i] = values[j]
						j++
					}
				}
				return nil, fmt.Errorf("TODO: %+v", values1)
			},
			Location: constructor.Location,
		}, nil
	}

	return constructor, nil
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

func structArgsInputFn(typ reflect.Type, expectNumValues int) func([]reflect.Value) (reflect.Value, error) {
	numFields := typ.NumField()
	return func(values []reflect.Value) (reflect.Value, error) {
		if len(values) != expectNumValues {
			return reflect.Value{}, fmt.Errorf("unexpected error, expected %d parameters got %d", numFields, len(values))
		}

		j := 0
		res := reflect.New(typ)
		for i := 0; i < numFields; i++ {
			f := typ.Field(i)
			if f.Type.AssignableTo(isStructArgsType) {
				continue
			}

			res.Field(i).Set(values[j])
			j++
		}
		return res.Elem(), nil
	}
}

func makeStructArgs(typ reflect.Type, values []reflect.Value) (reflect.Value, int) {
	numFields := typ.NumField()
	j := 0
	res := reflect.Zero(typ)
	for i := 0; i < numFields; i++ {
		f := typ.Field(i)
		if f.Type.AssignableTo(isStructArgsType) {
			continue
		}

		res.Field(i).Set(values[j])
		j++
	}
	return res, j
}

func structArgsOutputFn(typ reflect.Type, expectNumValues int) func(reflect.Value) ([]reflect.Value, error) {
	numFields := typ.NumField()
	return func(value reflect.Value) ([]reflect.Value, error) {
		j := 0
		res := make([]reflect.Value, expectNumValues)
		for i := 0; i < numFields; i++ {
			f := typ.Field(i)
			if f.Type.AssignableTo(isStructArgsType) {
				continue
			}

			if j >= expectNumValues {
				return nil, fmt.Errorf("unexpected number of fields")
			}
			res[j] = value.Field(i)
			j++
		}
		return res, nil
	}
}

func makeInSplicer(typ reflect.Type, i int, n int) func([]reflect.Value) ([]reflect.Value, error) {
	inFunc := structArgsInputFn(typ, n)

	return func(values []reflect.Value) ([]reflect.Value, error) {
		before := values[:i]
		splice := values[i : i+n]
		after := values[i+n:]
		replace, err := inFunc(splice)
		if err != nil {
			return nil, err
		}

		res := append(before, replace)
		res = append(res, after...)
		return res, nil
	}
}

func makeOutSplicer(typ reflect.Type, i int, n int) func([]reflect.Value) ([]reflect.Value, error) {
	inFunc := structArgsInputFn(typ, n)

	return func(values []reflect.Value) ([]reflect.Value, error) {
		before := values[:i]
		splice := values[i : i+n]
		after := values[i+n:]
		replace, err := inFunc(splice)
		if err != nil {
			return nil, err
		}

		res := append(before, replace)
		res = append(res, after...)
		return res, nil
	}
}
