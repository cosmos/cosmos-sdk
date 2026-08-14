package flag

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/client/v2/internal/util"
	"cosmossdk.io/math"
)

var isJSONFileRegex = regexp.MustCompile(`\.json$`)

type jsonMessageFlagType struct {
	messageDesc protoreflect.MessageDescriptor
}

func (j jsonMessageFlagType) NewValue(_ *context.Context, builder *Builder) Value {
	return &jsonMessageFlagValue{
		messageType:          util.ResolveMessageType(builder.TypeResolver, j.messageDesc),
		jsonMarshalOptions:   protojson.MarshalOptions{Resolver: builder.TypeResolver},
		jsonUnmarshalOptions: protojson.UnmarshalOptions{Resolver: builder.TypeResolver},
	}
}

func (j jsonMessageFlagType) DefaultValue() string {
	return ""
}

type jsonMessageFlagValue struct {
	jsonMarshalOptions   protojson.MarshalOptions
	jsonUnmarshalOptions protojson.UnmarshalOptions
	messageType          protoreflect.MessageType
	message              proto.Message
}

func (j *jsonMessageFlagValue) Get(protoreflect.Value) (protoreflect.Value, error) {
	if j.message == nil {
		return protoreflect.Value{}, nil
	}
	return protoreflect.ValueOfMessage(j.message.ProtoReflect()), nil
}

func (j *jsonMessageFlagValue) String() string {
	if j.message == nil {
		return ""
	}

	bz, err := j.jsonMarshalOptions.Marshal(j.message)
	if err != nil {
		return err.Error()
	}
	return string(bz)
}

func (j *jsonMessageFlagValue) Set(s string) error {
	j.message = j.messageType.New().Interface()
	var messageBytes []byte
	if isJSONFileRegex.MatchString(s) {
		jsonFile, err := os.Open(s)
		if err != nil {
			return err
		}
		messageBytes, err = io.ReadAll(jsonFile)
		if err != nil {
			return err
		}
	} else {
		messageBytes = []byte(s)
	}
	messageBytes, err := normalizeLegacyDecJSON(
		j.messageType.Descriptor(),
		messageBytes,
	)
	if err != nil {
		return err
	}
	return j.jsonUnmarshalOptions.Unmarshal(messageBytes, j.message)
}

func (j *jsonMessageFlagValue) Type() string {
	return fmt.Sprintf("%s (json)", j.messageType.Descriptor().FullName())
}

func normalizeLegacyDecJSON(
	descriptor protoreflect.MessageDescriptor,
	messageBytes []byte,
) ([]byte, error) {
	var value map[string]json.RawMessage
	if err := json.Unmarshal(messageBytes, &value); err != nil {
		return nil, err
	}

	if err := normalizeLegacyDecFields(descriptor, value); err != nil {
		return nil, err
	}

	return json.Marshal(value)
}

func normalizeLegacyDecFields(
	descriptor protoreflect.MessageDescriptor,
	value map[string]json.RawMessage,
) error {
	fields := descriptor.Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		fieldValue, ok := value[field.JSONName()]
		if !ok {
			fieldValue, ok = value[string(field.Name())]
		}
		if !ok {
			continue
		}

		normalized, err := normalizeLegacyDecField(field, fieldValue)
		if err != nil {
			return fmt.Errorf("invalid %s: %w", field.FullName(), err)
		}
		if _, hasJSONName := value[field.JSONName()]; hasJSONName {
			value[field.JSONName()] = normalized
		} else {
			value[string(field.Name())] = normalized
		}
	}

	return nil
}

func normalizeLegacyDecField(
	field protoreflect.FieldDescriptor,
	value json.RawMessage,
) (json.RawMessage, error) {
	if scalar, ok := GetScalarType(field); ok && scalar == DecScalarType {
		if field.IsList() {
			var values []json.RawMessage
			if err := json.Unmarshal(value, &values); err != nil {
				return nil, err
			}
			for i, item := range values {
				normalized, err := normalizeLegacyDecValue(item)
				if err != nil {
					return nil, err
				}
				values[i] = normalized
			}
			return json.Marshal(values)
		}

		return normalizeLegacyDecValue(value)
	}

	if field.Kind() != protoreflect.MessageKind {
		return value, nil
	}

	if field.IsMap() {
		var values map[string]json.RawMessage
		if err := json.Unmarshal(value, &values); err != nil {
			return nil, err
		}
		for key, item := range values {
			normalized, err := normalizeLegacyDecField(field.MapValue(), item)
			if err != nil {
				return nil, err
			}
			values[key] = normalized
		}
		return json.Marshal(values)
	}

	if field.IsList() {
		var values []json.RawMessage
		if err := json.Unmarshal(value, &values); err != nil {
			return nil, err
		}
		for i, item := range values {
			normalized, err := normalizeLegacyDecJSON(field.Message(), item)
			if err != nil {
				return nil, err
			}
			values[i] = normalized
		}
		return json.Marshal(values)
	}

	return normalizeLegacyDecJSON(field.Message(), value)
}

func normalizeLegacyDecValue(value json.RawMessage) (json.RawMessage, error) {
	if len(value) == 0 || value[0] != '"' {
		return value, nil
	}
	var text string
	if err := json.Unmarshal(value, &text); err != nil {
		return nil, err
	}
	// A cosmos.Dec field carries atomics: an integer already scaled by
	// LegacyPrecision. Only the decimal point distinguishes a human-readable
	// value from one that is encoded, so integer input is passed through
	// untouched rather than scaled a second time.
	if !strings.Contains(text, ".") {
		return value, nil
	}
	dec, err := math.LegacyNewDecFromStr(text)
	if err != nil {
		return nil, err
	}
	return json.Marshal(dec.BigInt().String())
}
