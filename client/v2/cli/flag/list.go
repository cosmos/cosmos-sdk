package flag

import (
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func bindSimpleListFlag(flagSet *pflag.FlagSet, kind protoreflect.Kind, name, shorthand, usage string) ListValue {
	switch kind {
	case protoreflect.StringKind:
		val := flagSet.StringSliceP(name, shorthand, nil, usage)
		return listValue(func(list protoreflect.List) {
			for _, x := range *val {
				list.Append(protoreflect.ValueOfString(x))
			}
		})
	case protoreflect.BytesKind:
		panic("TODO")
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
		protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		val := flagSet.UintSliceP(name, shorthand, nil, usage)
		return listValue(func(list protoreflect.List) {
			for _, x := range *val {
				list.Append(protoreflect.ValueOfUint64(uint64(x)))
			}
		})
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		val := flagSet.IntSliceP(name, shorthand, nil, usage)
		return listValue(func(list protoreflect.List) {
			for _, x := range *val {
				list.Append(protoreflect.ValueOfInt64(int64(x)))
			}
		})
	case protoreflect.BoolKind:
		val := flagSet.BoolSliceP(name, shorthand, nil, usage)
		return listValue(func(list protoreflect.List) {
			for _, x := range *val {
				list.Append(protoreflect.ValueOfBool(x))
			}
		})
	default:
		return nil
	}
}

type listValue func(protoreflect.List)

func (f listValue) AppendTo(list protoreflect.List) {
	f(list)
}

//type compositeListType struct {
//	simpleType Type
//}
//
//func (t compositeListType) NewValue(_ context.Context, _ *Options) pflag.Value {
//	return &compositeListValue{}
//}
//
//func (t compositeListType) DefaultValue() string {
//	return ""
//}
//
//type compositeListValue struct {
//	simpleType Type
//	values []protoreflect.Value
//	changed bool
//}
//
//func readAsCSV(val string) ([]string, error) {
//	if val == "" {
//		return []string{}, nil
//	}
//	stringReader := strings.NewReader(val)
//	csvReader := csv.NewReader(stringReader)
//	return csvReader.Read()
//}
//
//func writeAsCSV(vals []string) (string, error) {
//	b := &bytes.Buffer{}
//	w := csv.NewWriter(b)
//	err := w.Write(vals)
//	if err != nil {
//		return "", err
//	}
//	w.Flush()
//	return strings.TrimSuffix(b.String(), "\n"), nil
//}
//
//func (c compositeListValue) String() string {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (c *compositeListValue) Set(val string) error {
//	v, err := readAsCSV(val)
//	if err != nil {
//		return err
//	}
//	if !c.changed {
//		*c.values = v
//	} else {
//		*c.values = append(*c.values, v...)
//	}
//	c.changed = true
//	return nil
//}
//
//func (c compositeListValue) Type() string {
//	return fmt.Sprintf("slice of %s", c.simpleType)
//}
