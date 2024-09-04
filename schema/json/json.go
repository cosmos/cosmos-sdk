package schemajson

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"cosmossdk.io/schema"
	"cosmossdk.io/schema/addressutil"
)

var (
	trueBytes  = []byte("true")
	falseBytes = []byte("false")
	nullBytes  = []byte("null")
)

type encoder struct {
	addressCodec addressutil.AddressCodec
}

func (e encoder) marshalField(field schema.Field, value interface{}, writer io.Writer) error {
	if field.Nullable && value == nil {
		_, err := writer.Write(nullBytes)
		return err
	}

	switch field.Kind {
	case schema.BoolKind:
		value := value.(bool)
		if value {
			_, err := writer.Write(trueBytes)
			return err
		} else {
			_, err := writer.Write(falseBytes)
			return err
		}
	case schema.StringKind:
		panic("TODO deterministic string escaping")
	case schema.Uint8Kind, schema.Uint16Kind, schema.Uint32Kind,
		schema.Int8Kind, schema.Int16Kind, schema.Int32Kind:
		_, err := fmt.Fprintf(writer, "%d", value)
		return err
	case schema.Uint64Kind, schema.Int64Kind:
		_, err := fmt.Fprintf(writer, "\"%d\"", value)
		return err
	case schema.Float32Kind, schema.Float64Kind:
		_, err := fmt.Fprintf(writer, "%f", value)
		return err
	case schema.JSONKind:
		_, err := writer.Write(value.(json.RawMessage))
		return err
	case schema.EnumKind:
		_, err := fmt.Fprintf(writer, "\"%s\"", value)
		return err
	case schema.AddressKind:
		str, err := e.addressCodec.BytesToString(value.([]byte))
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(writer, "\"%s\"", str) // TODO escape
		return err
	case schema.BytesKind:
		value := value.([]byte)
		// TODO use AppendEncode and []byte directly??
		_, err := fmt.Fprintf(writer, "\"%s\"", base64.StdEncoding.EncodeToString(value))
		return err
	case schema.TimeKind:
		//value := value.(time.Time)
		//nanos := value.UnixNano()
		panic("TODO")
	case schema.DurationKind:
		panic("TODO")
	case schema.IntegerStringKind:
		panic("TODO")
	case schema.DecimalStringKind:
		panic("TODO")
	default:
		return fmt.Errorf("unsupported kind: %s", field.Kind)
	}
}
