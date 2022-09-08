package stablejson

import (
	"fmt"
	io "io"
	"sort"

	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	fieldsField      protoreflect.Name = "fields"
	valuesField      protoreflect.Name = "values"
	kindOneOf        protoreflect.Name = "kind"
	nullValueField   protoreflect.Name = "null_value"
	numberValueField protoreflect.Name = "number_value"
	stringValueField protoreflect.Name = "string_value"
	boolValueField   protoreflect.Name = "bool_value"
	structValueField protoreflect.Name = "struct_value"
	listValueField   protoreflect.Name = "list_value"
)

func marshalStruct(writer io.Writer, value protoreflect.Message) error {
	field := value.Descriptor().Fields().ByName(fieldsField)
	m1 := value.Get(field).Map()

	_, err := writer.Write([]byte("{"))
	if err != nil {
		return err
	}

	m2 := map[string]protoreflect.Message{}
	m1.Range(func(key protoreflect.MapKey, value protoreflect.Value) bool {
		m2[key.String()] = value.Message()
		return true
	})

	keys := maps.Keys(m2)
	sort.Strings(keys)
	first := true
	for _, k := range keys {
		if !first {
			_, err := writer.Write([]byte(","))
			if err != nil {
				return err
			}
		}

		first = false
		_, _ = fmt.Fprintf(writer, "%q:", k)

		err := marshalValue(writer, m2[k])
		if err != nil {
			return err
		}
	}

	_, err = writer.Write([]byte("}"))
	return err
}

func marshalListValue(writer io.Writer, value protoreflect.Message) error {
	field := value.Descriptor().Fields().ByName(valuesField)
	list := value.Get(field).List()
	n := list.Len()

	_, err := writer.Write([]byte("["))
	if err != nil {
		return err
	}

	first := true
	for i := 0; i < n; i++ {
		if !first {
			_, err = writer.Write([]byte(","))
			if err != nil {
				return err
			}
		}
		first = false

		err := marshalValue(writer, list.Get(i).Message())
		if err != nil {
			return err
		}
	}

	_, err = writer.Write([]byte("]"))
	return err
}

func marshalValue(writer io.Writer, value protoreflect.Message) error {
	field := value.WhichOneof(value.Descriptor().Oneofs().ByName(kindOneOf))
	if field == nil {
		return nil
	}

	var err error
	switch field.Name() {
	case nullValueField:
		_, err = writer.Write([]byte("null"))
	case numberValueField:
		err = marshalFloat(writer, value.Get(field).Float())
	case stringValueField:
		_, err = fmt.Fprintf(writer, "%q", value.Get(field).String())
	case boolValueField:
		_, err = fmt.Fprintf(writer, "%t", value.Get(field).Bool())
	case structValueField:
		return marshalStruct(writer, value.Get(field).Message())
	case listValueField:
		return marshalListValue(writer, value.Get(field).Message())
	default:
		return fmt.Errorf("unexpected field in google.protobuf.Value: %v", field)
	}
	return err
}
